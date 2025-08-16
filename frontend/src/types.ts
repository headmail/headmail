/**
 * Copyright 2025 JC-Lab
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

// Re-export types from generated API types for convenience
import type { components } from './generated/api-types';

// Type aliases for easier use
export type Campaign = components["schemas"]['domain.Campaign'];
export type List = components["schemas"]['domain.List'];
export type Subscriber = components["schemas"]['domain.Subscriber'];
export type Template = components["schemas"]['domain.Template'];
export type PaginationResponse = components["schemas"]['api_admin.PaginationResponse'];
export type PaginatedCampaignResponse = components["schemas"]['api_admin.PaginatedListResponse-domain_Campaign'];
export type PaginatedListResponse = components["schemas"]['api_admin.PaginatedListResponse-domain_List'];
export type PaginatedSubscriberResponse = components["schemas"]['api_admin.PaginatedListResponse-domain_Subscriber'];
export type PaginatedTemplateResponse = components["schemas"]['api_admin.PaginatedListResponse-domain_Template'];
