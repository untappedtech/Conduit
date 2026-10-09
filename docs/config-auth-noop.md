# No-Op Authorization

No-Op (No Operation) Authorization is Conduit's default and fallback security mode.

When active, Conduit performs **no authentication checks** on incoming requests. All clients are granted access according to the global permissions defined in your `policy` settings.

---

## When is No-Op Auth Active?

No-Op Auth is activated automatically when both `environmentEnabled` and `dbEnabled` are disabled or omitted:

```yaml
auth:
    environmentEnabled: false
    dbEnabled: false
```

It also serves as a safe fallback when no other authentication provider is successfully loaded.

---

## How Permissions Work in No-Op Mode

In No-Op mode, permissions are determined globally by the `policy` block in your configuration:

```yaml
policy:
    publicReads: true
    publicWrites: true
    publicMutation: true
```

- If `publicReads: true`, any client can issue `GET` and `HEAD` requests.
- If `publicWrites: true`, any client can insert, replace, update, or delete data rows (`POST`, `PUT`, `PATCH`, `DELETE`).
- If `publicMutation: true`, any client can create (`POST /v1/schema/<table>`) and drop (`DELETE /v1/schema/<table>`) database tables.

> **Warning:** If any policy flag is set to `false` while running No-Op Auth, operations in that category will be blocked (returning `401 Unauthorized`), because there is no authentication mechanism available to authenticate a user.

---

## Recommended Use Cases

- **Local Development:** Quick experimentation without having to generate, store, or transmit tokens.
- **Internal / Trusted Networks:** Microservices deployed inside private VPCs or behind an API Gateway / reverse proxy (such as Caddy, NGINX, or Envoy) that handles authentication upstream.
- **Air-Gapped Systems:** Standalone embedded databases on local machines.
