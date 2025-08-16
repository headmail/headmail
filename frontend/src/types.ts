/**
 * Copyright 2025 JC-Lab
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

// Re-export types from generated API types for convenience
import type { definitions } from './generated/api-types';

// Type aliases for easier use
export type Campaign = definitions['domain.Campaign'];
export type List = definitions['domain.List'];
export type Subscriber = definitions['domain.Subscriber'];
export type Template = definitions['domain.Template'];
export type PaginationResponse = definitions['api_admin.PaginationResponse'];
export type PaginatedCampaignResponse = definitions['api_admin.PaginatedListResponse-domain_Campaign'];
export type PaginatedListResponse = definitions['api_admin.PaginatedListResponse-domain_List'];
export type PaginatedSubscriberResponse = definitions['api_admin.PaginatedListResponse-domain_Subscriber'];
export type PaginatedTemplateResponse = definitions['api_admin.PaginatedListResponse-domain_Template'];
