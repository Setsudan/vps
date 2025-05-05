# Massive rework

## ✅ Done so far

1. **Models**

   * Core: `User`, `Message`
   * Legacy groups: `models/groups/Group`, `GroupMembership`
   * Friendships: `models/friendships/FriendRequest`
   * Guilds system: `models/guilds/Guild`, `GuildRole`, `GuildMember`, `Category`, `Channel`, `PermissionOverwrite`
   * Resumes: `models/resume/Resume` + submodels (Education, Experience, Project, Certification, Skill, Interest)

2. **Repositories**

   * `UserRepository`, `MessagingRepository`, `GroupRepository`, `GroupMembershipRepository`
   * `FriendRequestRepository`, `GuildRepository`, `GuildMemberRepository`
   * `CategoryRepository`, `ChannelRepository`, `PermissionOverwriteRepository`
   * `ResumeRepository`

3. **Services**

   * **Auth** (register/login/JWT)
   * **Users** (avatar, list, get, update profile)
   * **Legacy Groups** (create/list/update/delete + membership)
   * **Friendships** (send/respond/list requests & friends)
   * **Messaging** (Redis-backed send, reactions, history, transfer)
   * **Presence** (online/offline via Redis TTL + WS)
   * **Resumes** (CRUD per user)
   * **Guilds** (create/list/update/delete + membership roles)

4. **Controllers**

   * **Helpers** (`parseJWT`, `buildUpgrader`)
   * **AuthController**, **UserController**, **MessagingController**, **PresenceController**
   * **GroupController** (legacy groups)
   * **FriendshipController**, **ResumeController**, **GuildController**

5. **Boot & Routing**

   * `init_server.go` wired all repos, services, controllers, DB migrations, Redis listeners
   * `router.go` registers routes for health, auth, presence, users, messaging, groups, resumes, friendships, guilds

---

## 🏗️ Still to do

1. **Permissions**

   * Service & controller for `PermissionOverwrite` endpoints.

2. **Categories**

   * Service & controller for guild‐categories (CRUD).

3. **Channels**

   * Service & controller for text/audio channels (CRUD).

4. **Role Management**

   * Endpoints for `GuildRole` (create/list/update/delete) under `/guilds/:guild_id/roles`.

5. **Integration**

   * Hook the new permissions/categories/channels controllers into `init_server.go` and `router.go`.

6. **Messaging & Group Rework**

   * Adapt legacy messaging/group flows to respect friendships and guild permissions.

7. **Docs & Tests**

   * (Optional) Swagger/OpenAPI spec, unit & integration tests.
