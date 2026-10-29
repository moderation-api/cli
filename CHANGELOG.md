# Changelog

## 0.7.0 (2026-10-29)

Full Changelog: [v0.6.0...v0.7.0](https://github.com/moderation-api/cli/compare/v0.6.0...v0.7.0)

### API changes

- `POST /moderate`: added the new `hidden_lookalike` enum value to the request property `policies/items/oneOf[subschema #36: Unicode Spoofing]/signals/propertyNames/`
- `POST /moderate`: added the new `leetspeak` enum value to the request property `policies/items/oneOf[subschema #36: Unicode Spoofing]/signals/propertyNames/`
- `DELETE /wordlist/{id}`: endpoint added

## 0.6.0 (2026-10-28)

Full Changelog: [v0.5.0...v0.6.0](https://github.com/moderation-api/cli/compare/v0.5.0...v0.6.0)

### API changes

- `POST /actions/execute`: added the new optional request property `createCasebookCase`
- `POST /actions/execute`: added the new optional request property `moderator`
- `POST /authors/{id}/block`: endpoint added
- `POST /authors/{id}/suspend`: endpoint added
- `POST /authors/{id}/unblock`: endpoint added
- `POST /content/{id}/allow`: endpoint added
- `POST /content/{id}/reject`: endpoint added

## 0.5.0 (2026-10-26)

Full Changelog: [v0.4.0...v0.5.0](https://github.com/moderation-api/cli/compare/v0.4.0...v0.5.0)

### ⚠ Breaking changes

- `GET /authors`: removed the enum value `reportCount` from the `query` request parameter `sortBy`
- `GET /queue/{id}`: the response property `queue/description` became nullable for the status `200`
- `GET /queue/{id}`: removed the required property `queue/filter` from the response with the `200` status
- `GET /queue/{id}/items`: added `subschema #1, subschema #2` to the `data/items/content` response property `anyOf` list for the response status `200`
- `GET /queue/{id}/items`: the response property `data/items/authorId` became nullable for the status `200`
- `GET /queue/{id}/items`: the response property `data/items/contentType` became nullable for the status `200`
- `GET /queue/{id}/items`: the response property `data/items/conversationId` became nullable for the status `200`
- `GET /queue/{id}/items`: the response property `data/items/language` became nullable for the status `200`
- `GET /queue/{id}/items`: the response property `data/items/labels/items/flagged` became optional for the status `200`
- `GET /queue/{id}/items`: the `data/items/content` response's property `type` changed from `string` to `any` for status `200`
- `POST /queue/{id}/items/{itemId}/resolve`: removed the request body
- `POST /queue/{id}/items/{itemId}/unresolve`: removed the request body
- `GET /queue/{id}/stats`: the response property `topReviewers/items/name` became nullable for the status `200`
- `GET /queue/{id}/stats`: the response property `topReviewers/items/userId` became nullable for the status `200`
- `GET /queue/{id}/stats`: removed the required property `actionStats` from the response with the `200` status
- `GET /queue/{id}/stats`: removed the required property `reviewStats/averageTimeToReview` from the response with the `200` status
- `GET /queue/{id}/stats`: removed the required property `topReviewers/items/averageTimePerReview` from the response with the `200` status
- `GET /queue/{id}/stats`: removed the required property `topReviewers/items/topActions` from the response with the `200` status
- `GET /queue/{id}/stats`: removed the required property `trends/flaggedContentTrends` from the response with the `200` status

### API changes

- `GET /actions`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `POST /actions/execute`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `POST /actions/{actionId}/execute`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `DELETE /actions/{id}`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `GET /actions/{id}`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `PUT /actions/{id}`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `GET /auth`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `GET /authors`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `DELETE /authors/{id}`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `GET /authors/{id}`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `PUT /authors/{id}`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `DELETE /channels/{id}`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `GET /channels/{id}`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `PUT /channels/{id}`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `POST /channels/{id}/duplicate`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `DELETE /projects/{id}`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `GET /projects/{id}`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `PUT /projects/{id}`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `POST /projects/{id}/api-keys`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `POST /projects/{id}/duplicate`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `GET /projects/{projectId}/channels`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `POST /projects/{projectId}/channels`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `GET /queue/{id}`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `GET /queue/{id}/items`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `GET /queue/{id}/stats`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `DELETE /webhooks/{id}`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `GET /webhooks/{id}`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `PUT /webhooks/{id}`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `GET /wordlist`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `GET /wordlist/{id}`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `PUT /wordlist/{id}`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `GET /wordlist/{id}/embedding-status`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `DELETE /wordlist/{id}/words`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `POST /wordlist/{id}/words`: added the new `queue_item_not_found` enum value to the `code` response property for the response status `404`
- `GET /authors`: deleted the `query` request parameter `contentTypes`
- `GET /queue/{id}/items`: deleted the `query` request parameter `sortDirection`
- `GET /queue/{id}/items`: deleted the `query` request parameter `sortField`
- a breaking change was detected but the version is still `1.1.0`
- `GET /analytics/authors`: endpoint added
- `GET /analytics/automation`: endpoint added
- `GET /analytics/channels`: endpoint added
- `GET /analytics/policies`: endpoint added
- `GET /analytics/summary`: endpoint added
- `GET /analytics/volumes`: endpoint added
- `POST /feedback`: endpoint added
- `GET /queue`: endpoint added
- `GET /queue/{id}`: the `queue/resolvedItemsCount` response's property `type` was narrowed from `number` to `integer` for status `200`
- `GET /queue/{id}`: the `queue/totalItemsCount` response's property `type` was narrowed from `number` to `integer` for status `200`
- `GET /queue/{id}`: the `queue/unresolvedItemsCount` response's property `type` was narrowed from `number` to `integer` for status `200`
- `GET /queue/{id}`: added the required property `queue/isDefault` to the response with the `200` status
- `GET /queue/{id}/items`: added the new optional `query` request parameter `sort`
- `GET /queue/{id}/items`: added the optional property `data/items/labels/items/manual` to the response with the `200` status
- `GET /queue/{id}/items`: added the optional property `data/items/labels/items/matches` to the response with the `200` status
- `GET /queue/{id}/items`: added the optional property `data/items/labels/items/shadow_flagged` to the response with the `200` status
- `GET /queue/{id}/items`: removed the optional property `data/items/actions` from the response with the `200` status
- `GET /queue/{id}/items`: the response property `data/items/authorId` became required for the status `200`
- `GET /queue/{id}/items`: the response property `data/items/contentType` became required for the status `200`
- `GET /queue/{id}/items`: the response property `data/items/conversationId` became required for the status `200`
- `GET /queue/{id}/items`: the response property `data/items/language` became required for the status `200`
- `GET /queue/{id}/items`: added the required property `data/items/contentId` to the response with the `200` status
- `GET /queue/{id}/items`: added the required property `data/items/severity` to the response with the `200` status
- `POST /queue/{id}/items/resolve`: endpoint added
- `POST /queue/{id}/items/{itemId}/resolve`: removed the optional property `comment` from the response with the `200` status
- `POST /queue/{id}/items/{itemId}/unresolve`: the `status` response property for the response status `200` was restricted to a list of enum values
- `GET /queue/{id}/stats`: removed the optional property `topReviewers/items/accuracyScore` from the response with the `200` status
- `GET /queue/{id}/stats`: the `reviewStats/totalPending` response's property `type` was narrowed from `number` to `integer` for status `200`
- `GET /queue/{id}/stats`: the `reviewStats/totalReviewed` response's property `type` was narrowed from `number` to `integer` for status `200`
- `GET /queue/{id}/stats`: the `topReviewers/items/reviewCount` response's property `type` was narrowed from `number` to `integer` for status `200`
- `GET /queue/{id}/stats`: the `trends/dailyReviewCounts/items/count` response's property `type` was narrowed from `number` to `integer` for status `200`
- `POST /wordlist`: endpoint added

## 0.4.0 (2026-10-23)

Full Changelog: [v0.3.3...v0.4.0](https://github.com/moderation-api/cli/compare/v0.3.3...v0.4.0)

### API changes

- `POST /projects/{id}/api-keys`: endpoint added

## 0.3.3 (2026-10-23)

Full Changelog: [v0.3.2...v0.3.3](https://github.com/moderation-api/cli/compare/v0.3.2...v0.3.3)

### Chores

- Update the generated code.

## 0.3.2 (2026-10-23)

Full Changelog: [v0.3.1...v0.3.2](https://github.com/moderation-api/cli/compare/v0.3.1...v0.3.2)

### Chores

- Update the generated code.

## 0.3.1 (2026-10-23)

Full Changelog: [v0.3.0...v0.3.1](https://github.com/moderation-api/cli/compare/v0.3.0...v0.3.1)

### Chores

- Update the generated code.

## 0.3.0 (2026-10-22)

Full Changelog: [v0.2.1...v0.3.0](https://github.com/moderation-api/cli/compare/v0.2.1...v0.3.0)

### ⚠ Breaking changes

- `GET /actions`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `POST /actions`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `POST /actions/execute`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `POST /actions/{actionId}/execute`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `DELETE /actions/{id}`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `GET /actions/{id}`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `PUT /actions/{id}`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `GET /auth`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `POST /auth`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `GET /authors`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `POST /authors`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `DELETE /authors/{id}`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `GET /authors/{id}`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `PUT /authors/{id}`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `DELETE /channels/{id}`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `PUT /channels/{id}`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `POST /channels/{id}/duplicate`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `POST /moderate`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `POST /moderate/image`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `POST /moderate/text`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `POST /projects`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `PUT /projects/{id}`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `POST /projects/{id}/duplicate`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `POST /projects/{projectId}/channels`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `GET /queue/{id}`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `GET /queue/{id}/items`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `POST /queue/{id}/items/{itemId}/resolve`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `POST /queue/{id}/items/{itemId}/unresolve`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `GET /queue/{id}/stats`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `POST /webhooks`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `PUT /webhooks/{id}`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `GET /wordlist`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `GET /wordlist/{id}`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `PUT /wordlist/{id}`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `GET /wordlist/{id}/embedding-status`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `DELETE /wordlist/{id}/words`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `POST /wordlist/{id}/words`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`

### API changes

- a breaking change was detected but the version is still `1.1.0`

## 0.2.1 (2026-10-22)

Full Changelog: [v0.2.0...v0.2.1](https://github.com/moderation-api/cli/compare/v0.2.0...v0.2.1)

### Chores

- Update the generated code.

## 0.2.0 (2026-10-22)

Full Changelog: [v0.1.0...v0.2.0](https://github.com/moderation-api/cli/compare/v0.1.0...v0.2.0)

### API changes

- Update to the latest API spec.

## 0.1.0 (2026-10-22)

The first release of `modapi`, the Moderation API command-line interface.

```bash
curl -fsSL https://raw.githubusercontent.com/moderation-api/cli/main/scripts/install.sh | bash
```

- Every API endpoint as a command, for example `modapi content submit` and `modapi authors list`.
- Reads your key from `MODAPI_SECRET_KEY`, the OS keychain (`modapi configure`) or `--secret-key`.
- JSON output and `--jq` filters for scripts, `--all` to fetch every page, and `--dry-run` to see a request without sending it.
- Built for AI coding agents: `--usage` and `--schema` describe every command and request body.
