# Discovery Answers

## Q1: Will this MCP server be deployed as a standalone service that AI assistants connect to remotely?
**Answer:** Yes - SSE transport with X-OCM-OFFLINE-TOKEN header support

## Q2: Should the server support multiple concurrent OCM authentication tokens from different users?
**Answer:** Yes

## Q3: Will users need to create ROSA HCP clusters in different AWS regions beyond us-east-1?
**Answer:** Yes, but us-east-1 should be the default if no region is specified

## Q4: Should the server cache cluster information to reduce OCM API calls?
**Answer:** No

## Q5: Will this server need to handle ROSA HCP cluster lifecycle beyond creation (scaling, upgrading, deletion)?
**Answer:** Not for the MVP