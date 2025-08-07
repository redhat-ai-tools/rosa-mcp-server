# Discovery Questions

## Q1: Will this MCP server be deployed as a standalone service that AI assistants connect to remotely?
**Default if unknown:** Yes (most MCP servers run as independent services for multiple AI assistant connections)

## Q2: Should the server support multiple concurrent OCM authentication tokens from different users?
**Default if unknown:** No (MVP should focus on single-user authentication for simplicity)

## Q3: Will users need to create ROSA HCP clusters in different AWS regions beyond us-east-1?
**Default if unknown:** Yes (multi-region support is standard for production ROSA deployments)

## Q4: Should the server cache cluster information to reduce OCM API calls?
**Default if unknown:** No (real-time data accuracy is more important than performance for an MVP)

## Q5: Will this server need to handle ROSA HCP cluster lifecycle beyond creation (scaling, upgrading, deletion)?
**Default if unknown:** No (MVP focuses on essential operations - creation and monitoring)