# ChefBook Backend API Architecture

This document is a working map of the backend API as it exists in the repository today.

## Purpose

- explain where the external API lives
- show which service owns which domain
- show how services call each other
- give a stable reference for future backend and client work

## Top-Level Shape

```mermaid
flowchart LR
    Client["Mobile client / web client"] --> Gateway["API Gateway<br/>HTTP REST /v1"]

    Gateway --> Auth["Auth Service<br/>gRPC"]
    Gateway --> User["User Service<br/>gRPC"]
    Gateway --> Profile["Profile Service<br/>gRPC"]
    Gateway --> Tag["Tag Service<br/>gRPC"]
    Gateway --> Recipe["Recipe Service<br/>gRPC"]
    Gateway --> Encryption["Encryption Service<br/>gRPC"]
    Gateway --> Shopping["Shopping List Service<br/>gRPC"]
    Gateway --> Subscription["Subscription Service<br/>gRPC<br/>(external to this repo)"]

    Auth --> AuthDB["PostgreSQL"]
    User --> UserDB["PostgreSQL"]
    Profile --> ProfileDeps["Remote repositories / aggregators"]
    Tag --> TagDB["PostgreSQL"]
    Recipe --> RecipeDB["PostgreSQL"]
    Recipe --> S3["S3 / object storage"]
    Encryption --> EncryptionDB["PostgreSQL"]
    Shopping --> ShoppingDB["PostgreSQL"]

    Auth -. optional MQ .- MQ["AMQP / Message Queue"]
    User -. optional MQ .- MQ
    Recipe -. MQ pub/sub .- MQ
    Encryption -. MQ pub/sub .- MQ
    Shopping -. optional MQ .- MQ
```

## Request Path

1. Client calls `api-gateway` over HTTP.
2. Gateway applies recovery, request logging, and rate limiting.
3. Protected endpoints pass through auth middleware.
4. Auth middleware fetches the JWT public key from `AuthService` and caches it for a refresh interval.
5. Gateway handlers translate HTTP requests into gRPC calls to domain services.
6. Domain services execute business logic using their own database and, where needed, S3, MQ, or other services.

## API Gateway

Gateway entrypoint:
- `chefbook-backend/api-gateway`

Gateway responsibilities:
- public HTTP API under `/v1`
- JWT validation and user authorization
- request logging and rate limiting
- HTTP-to-gRPC translation
- Swagger docs in non-release mode
- health endpoint at `/healthz`

Global middleware:
- `gin.Recovery()`
- request log middleware
- rate limiter
- auth middleware on protected routes

Docs and health:
- `/doc/*any`
- `/healthz`

## External HTTP API Groups

Base prefix:
- `/v1`

Route groups:
- `/auth` for sign-up, activation, sign-in, refresh, sign-out, OAuth, sessions, password flows, nickname flows
- `/subscriptions` for subscription reads and Google subscription confirmation
- `/profile` for current user profile data and avatar management
- `/profiles/:profileId` for reading another profile
- `/recipes` for recipe CRUD, book, favourites, pictures, rating, translations, and recipe-to-collection binding
- `/recipes/tags` for tags lookup inside recipe flows
- `/collections` for collection CRUD and save/remove from recipe book
- `/encryption/vault` for encrypted vault lifecycle
- `/encryption/recipes/:recipeId` for recipe key ownership and sharing
- `/shopping-lists` for personal/shared shopping list flows, users, and invite links

## gRPC Service Ownership

### `AuthService`

Owns:
- account lifecycle
- JWT issuing
- OAuth integration
- sessions
- password reset / change
- nickname availability and assignment
- profile deletion state

Main RPC families:
- `SignUp`, `ActivateProfile`, `SignIn`, `RefreshSession`, `SignOut`
- `RequestGoogleOAuth`, `SignInGoogle`, `ConnectGoogle`, `DeleteGoogleConnection`
- `RequestVkOAuth`, `SignInVk`, `ConnectVk`, `DeleteVkConnection`
- `GetSessions`, `EndSessions`
- `RequestPasswordReset`, `ResetPassword`, `ChangePassword`
- `GetProfileDeletionStatus`, `DeleteProfile`, `CancelProfileDeletion`
- `GetAccessTokenPublicKey`, `GetAuthInfo`, `GetVisibleNames`, `CheckNicknameAvailability`, `SetNickname`

### `UserService`

Owns:
- source-of-truth user social data
- name
- description
- avatar upload lifecycle

Main RPC families:
- `GetUsersMinInfo`, `GetUserInfo`
- `SetUserName`, `SetUserDescription`
- `GenerateUserAvatarUploadLink`, `ConfirmUserAvatarUploading`, `DeleteUserAvatar`

### `ProfileService`

Owns:
- aggregated profile read models

Main RPC families:
- `GetProfile`
- `GetProfilesMinInfo`

Important note:
- gateway profile handlers are composite and use `AuthService`, `UserService`, and `ProfileService` together

### `TagService`

Owns:
- tag taxonomy and lookup

Main RPC families:
- `GetTags`
- `GetTagsMap`
- `GetTag`
- `GetTagGroups`

### `RecipeService`

Owns:
- recipes
- recipe book state
- favourites
- collections
- recipe pictures upload coordination
- translations
- rating
- recipe policy reads

Main RPC families:
- `GetRecipes`, `GetRandomRecipe`, `GetRecipeBook`
- `CreateRecipe`, `GetRecipe`, `UpdateRecipe`, `DeleteRecipe`
- `GenerateRecipePicturesUploadLinks`, `SetRecipePictures`
- `RateRecipe`
- `SaveRecipeToRecipeBook`, `RemoveRecipeFromRecipeBook`
- `SaveRecipeToFavourites`, `RemoveRecipeFromFavourites`
- `AddRecipeToCollection`, `RemoveRecipeFromCollection`, `SetRecipeCollections`
- `TranslateRecipe`, `DeleteRecipeTranslation`
- `GetRecipePolicy`, `GetRecipeNames`
- `GetCollections`, `CreateCollection`, `GetCollection`, `UpdateCollection`, `DeleteCollection`
- `SaveCollectionToRecipeBook`, `RemoveCollectionFromRecipeBook`

### `EncryptionService`

Owns:
- encrypted vault metadata
- recipe key storage and sharing requests

Main RPC families:
- `HasEncryptedVault`, `GetEncryptedVaultKey`, `CreateEncryptedVault`
- `RequestEncryptedVaultDeletion`, `DeleteEncryptedVault`
- `GetRecipeKeyRequests`, `RequestRecipeKeyAccess`
- `GetRecipeKey`, `SetRecipeKey`, `DeleteRecipeKey`

### `ShoppingListService`

Owns:
- personal and shared shopping lists
- shopping list user membership
- shopping list invite link flow

Main RPC families:
- `GetShoppingLists`
- `CreateSharedShoppingList`, `GetShoppingList`
- `SetShoppingListName`, `SetShoppingList`, `AddPurchasesToShoppingList`
- `DeleteSharedShoppingList`
- `GetShoppingListUsers`, `GetSharedShoppingListLink`
- `JoinShoppingList`, `DeleteUserFromShoppingList`

### `SubscriptionService`

Status:
- referenced by gateway and some services
- implementation is not present in this repository snapshot

Known roles from code:
- subscription reads in gateway
- subscription-aware checks in recipe, encryption, shopping-list
- remote dependency of auth service

## Inter-Service Dependency Graph

```mermaid
flowchart LR
    Gateway["API Gateway"] --> Auth
    Gateway --> User
    Gateway --> Profile
    Gateway --> Tag
    Gateway --> Recipe
    Gateway --> Encryption
    Gateway --> Shopping["Shopping List"]
    Gateway --> Subscription

    Auth --> Subscription

    Recipe --> Profile
    Recipe --> Tag
    Recipe --> Encryption

    Encryption --> Auth
    Encryption --> Profile
    Encryption --> Recipe

    Shopping --> Profile
    Shopping --> Recipe
```

## Storage and Integration Boundaries

Per-service storage:
- `auth` uses PostgreSQL
- `user` uses PostgreSQL
- `tag` uses PostgreSQL
- `recipe` uses PostgreSQL
- `encryption` uses PostgreSQL
- `shopping-list` uses PostgreSQL

Extra integrations:
- `recipe` uses S3 for recipe pictures
- `auth`, `user`, `recipe`, `encryption`, `shopping-list` include AMQP integration paths
- all main services expose gRPC and gRPC health checks

## Database Schema

Each service owns its own PostgreSQL schema. Foreign keys are local to a service database only; IDs that point to another service, such as `owner_id`, `user_id`, or `recipe_id`, are logical cross-service references and are not enforced by PostgreSQL.

### Auth Database

Owns credentials, sessions, OAuth bindings, Firebase bindings, activation codes, password reset codes, profile deletion requests, and outgoing events.

```mermaid
erDiagram
    AUTH_USERS {
        uuid user_id PK
        varchar email UK
        varchar nickname UK
        varchar password
        role role
        boolean activated
        timestamptz deletion_timestamp
    }

    ACTIVATION_CODES {
        uuid user_id FK,UK
        varchar activation_code
    }

    SESSIONS {
        serial session_id PK
        uuid user_id FK
        varchar refresh_token UK
        inet ip
        text user_agent
        timestamptz last_access
        timestamptz expires_at
    }

    PASSWORD_RESETS {
        uuid user_id FK
        varchar reset_code UK
        boolean used
        timestamptz expires_at
    }

    OAUTH {
        uuid user_id FK,UK
        text google_id UK
        bigint vk_id UK
    }

    FIREBASE {
        uuid user_id FK,UK
        text firebase_id UK
    }

    DELETE_PROFILE_REQUESTS {
        uuid user_id FK,UK
        boolean with_shared_data
        timestamptz deletion_timestamp
    }

    AUTH_OUTBOX {
        uuid message_id PK
        varchar exchange
        varchar type
        jsonb body
    }

    AUTH_USERS ||--o| ACTIVATION_CODES : owns
    AUTH_USERS ||--o{ SESSIONS : owns
    AUTH_USERS ||--o{ PASSWORD_RESETS : owns
    AUTH_USERS ||--o| OAUTH : owns
    AUTH_USERS ||--o| FIREBASE : owns
    AUTH_USERS ||--o| DELETE_PROFILE_REQUESTS : owns
```

Important constraints:
- `users.email` and `users.nickname` are unique.
- `activation_codes`, `oauth`, `firebase`, and `delete_profile_requests` are one-to-one with `users`.
- `password_resets.reset_code` and `sessions.refresh_token` are globally unique.

### User Database

Owns public profile fields and avatar upload lifecycle. Auth identity is linked by the same `user_id`, but there is no database FK to the auth service.

```mermaid
erDiagram
    USER_USERS {
        uuid user_id PK
        varchar first_name
        varchar last_name
        varchar description
        uuid avatar_id
    }

    AVATAR_UPLOADS {
        uuid avatar_id PK
        uuid user_id FK,UK
    }

    USER_INBOX {
        uuid message_id PK
        timestamptz timestamp
    }

    USER_USERS ||--o| AVATAR_UPLOADS : has_pending_upload
```

### Tag Database

Owns localized tag groups and tags used by recipe search and recipe metadata.

```mermaid
erDiagram
    TAG_GROUPS {
        varchar group_id PK
        varchar name_en
        varchar name_ru
    }

    TAGS {
        varchar tag_id PK
        varchar name_en
        varchar name_ru
        varchar emoji
        varchar group_id FK
    }

    TAG_GROUPS ||--o{ TAGS : contains
```

### Recipe Database

Owns recipes, recipe read state, favourites, scores, translations, pictures, collections, collection membership, collection sharing keys, and MQ inbox/outbox.

```mermaid
erDiagram
    RECIPES {
        uuid recipe_id PK
        varchar name
        uuid owner_id
        visibility visibility
        boolean encrypted
        varchar language
        text_array translations
        jsonb ingredients
        jsonb cooking
        jsonb pictures
        decimal rating
        int votes
        text_array tags
        int version
    }

    TRANSLATIONS {
        uuid recipe_id FK
        varchar language
        uuid author_id
        varchar name
        jsonb ingredients
        jsonb cooking
    }

    RECIPE_PICTURES_UPLOADS {
        uuid recipe_id FK,UK
        jsonb pictures
    }

    RECIPE_BOOK {
        uuid recipe_id FK
        uuid user_id
    }

    FAVOURITES {
        uuid recipe_id FK
        uuid user_id
    }

    SCORES {
        uuid recipe_id FK
        uuid user_id
        smallint score
    }

    COLLECTIONS {
        uuid collection_id PK
        varchar name
        visibility visibility
    }

    COLLECTIONS_KEYS {
        uuid collection_id FK,UK
        uuid key
        timestamptz expires_at
    }

    COLLECTIONS_CONTRIBUTORS {
        uuid collection_id FK
        uuid contributor_id
        contributor_role role
    }

    COLLECTIONS_USERS {
        uuid collection_id FK
        uuid user_id
    }

    RECIPES_COLLECTIONS {
        uuid recipe_id FK
        uuid collection_id FK
        timestamptz binding_timestamp
    }

    RECIPE_INBOX {
        uuid message_id PK
        timestamptz timestamp
    }

    RECIPE_OUTBOX {
        uuid message_id PK
        varchar exchange
        varchar type
        jsonb body
    }

    RECIPES ||--o{ TRANSLATIONS : has
    RECIPES ||--o| RECIPE_PICTURES_UPLOADS : has_upload_state
    RECIPES ||--o{ RECIPE_BOOK : saved_by
    RECIPES ||--o{ FAVOURITES : favourited_by
    RECIPES ||--o{ SCORES : rated_by
    RECIPES ||--o{ RECIPES_COLLECTIONS : assigned_to
    COLLECTIONS ||--o| COLLECTIONS_KEYS : has_invite_key
    COLLECTIONS ||--o{ COLLECTIONS_CONTRIBUTORS : editable_by
    COLLECTIONS ||--o{ COLLECTIONS_USERS : saved_by
    COLLECTIONS ||--o{ RECIPES_COLLECTIONS : contains
```

Important constraints:
- `recipes.encrypted=true` cannot be combined with `visibility='public'`.
- `translations` are unique by `(recipe_id, language, author_id)`.
- `recipe_book`, `favourites`, `scores`, `collections_contributors`, `collections_users`, and `recipes_collections` are unique by their pair keys.
- `recipes.translations` and `recipes.tags` have GIN indexes.
- `owner_id`, `user_id`, `author_id`, and `contributor_id` reference users logically across services.

### Encryption Database

Owns encrypted vault keys, recipe access keys, key request status, vault deletion codes, and MQ inbox/outbox.

```mermaid
erDiagram
    VAULT_KEYS {
        uuid user_id PK
        text public_key
        text private_key
        text salt
    }

    RECIPE_KEYS {
        uuid recipe_id
        uuid user_id FK
        text key
        recipe_key_request_status status
    }

    VAULT_DELETIONS {
        uuid user_id FK,UK
        varchar delete_code
    }

    ENCRYPTION_INBOX {
        uuid message_id PK
        timestamptz timestamp
    }

    ENCRYPTION_OUTBOX {
        uuid message_id PK
        varchar exchange
        varchar type
        jsonb body
    }

    VAULT_KEYS ||--o{ RECIPE_KEYS : owns_or_requests
    VAULT_KEYS ||--o| VAULT_DELETIONS : pending_deletion
```

Important constraints:
- `recipe_keys` are unique by `(recipe_id, user_id)`.
- `recipe_keys.recipe_id` is a logical reference to the recipe service.
- `recipe_keys.user_id` is a local FK to `vault_keys`, not to auth/user.

### Shopping List Database

Owns personal/shared shopping lists, per-user list names, invite keys, and MQ inbox.

```mermaid
erDiagram
    SHOPPING_LISTS {
        uuid shopping_list_id PK
        shopping_list_type type
        uuid owner_id
        jsonb purchases
        int version
    }

    SHOPPING_LISTS_USERS {
        uuid shopping_list_id FK
        uuid user_id
        varchar name
    }

    SHOPPING_KEYS {
        uuid shopping_list_id FK,UK
        uuid key
        timestamptz expires_at
    }

    SHOPPING_INBOX {
        uuid message_id PK
        timestamptz timestamp
    }

    SHOPPING_LISTS ||--o{ SHOPPING_LISTS_USERS : visible_to
    SHOPPING_LISTS ||--o| SHOPPING_KEYS : has_invite_key
```

Important constraints:
- `shopping_lists_users` is unique by `(shopping_list_id, user_id)`.
- `shopping_lists_users.name` is the user's display name for that list, not the profile/user name.
- `shopping_lists.version` is used for optimistic updates.
- `owner_id` and `shopping_lists_users.user_id` are logical cross-service user references.

## Internal Service Template

Most services follow this shape:

- `cmd/` for runtime entrypoints and migrations
- `internal/app` for process bootstrap
- `internal/transport/grpc` for gRPC server adapters
- `internal/repository/postgres` for persistence
- `internal/repository/grpc` for remote service clients when needed
- `migrations/` for schema changes
- `deployments/helm` for Kubernetes deployment
- `api/proto/contract/v1` for service contract source
- `api/proto/implementation/v1` for generated gRPC code

## Practical Rules For Future Work

- If the task changes public HTTP shape, start in `api-gateway`.
- If the task changes domain behavior, start in the owning service.
- If the task changes a gRPC contract, update both the owning service and every caller.
- If the task touches profile endpoints, expect orchestration across `auth`, `user`, and `profile`.
- If the task touches encrypted recipes, expect `recipe` and `encryption` to be coupled.
- If the task touches shopping lists, check both direct list logic and recipe/profile dependencies.
- If the task touches subscriptions, verify whether the dependency lives outside this monorepo.

## Important Gaps And Caveats

- `SubscriptionService` is referenced but not implemented in this repository.
- `ProfileService` is an aggregator; some profile behavior is split between gateway orchestration and downstream services.
- MQ event topology is present in code, but queue names and event contracts were not mapped in this first pass.
- Swagger exists for the gateway, but this document is the clearer architecture map for human navigation.

## Recommended Next Artifacts

- a route-to-RPC matrix for every `/v1` endpoint
- an event/MQ topology document
- a data ownership table per service and table family
