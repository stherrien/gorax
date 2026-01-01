import { WorkflowDefinition } from '../api/workflows'

// Marketplace template type
export interface MarketplaceTemplate {
  id: string
  name: string
  description: string
  category: string
  definition: WorkflowDefinition
  tags: string[]
  authorId: string
  authorName: string
  version: string
  downloadCount: number
  averageRating: number
  totalRatings: number
  isVerified: boolean
  sourceTenantId?: string
  sourceTemplateId?: string
  publishedAt: string
  updatedAt: string
}

// Template category type
export type TemplateCategory =
  | 'security'
  | 'monitoring'
  | 'integration'
  | 'dataops'
  | 'devops'
  | 'notification'
  | 'automation'
  | 'analytics'
  | 'other'

// Template review type
export interface TemplateReview {
  id: string
  templateId: string
  tenantId: string
  userId: string
  userName: string
  rating: number
  comment: string
  createdAt: string
  updatedAt: string
}

// Template installation type
export interface TemplateInstallation {
  id: string
  templateId: string
  tenantId: string
  userId: string
  workflowId: string
  installedVersion: string
  installedAt: string
}

// Publish template input
export interface PublishTemplateInput {
  name: string
  description: string
  category: string
  definition: WorkflowDefinition
  tags?: string[]
  version: string
  sourceTemplateId?: string
}

// Install template input
export interface InstallTemplateInput {
  workflowName: string
}

// Install template result
export interface InstallTemplateResult {
  workflowId: string
  workflowName: string
  definition: WorkflowDefinition
}

// Rate template input
export interface RateTemplateInput {
  rating: number
  comment?: string
}

// Search filter
export interface SearchFilter {
  category?: string
  tags?: string[]
  searchQuery?: string
  minRating?: number
  isVerified?: boolean
  sortBy?: 'popular' | 'recent' | 'rating'
  page?: number
  limit?: number
}

// Marketplace list response
export interface MarketplaceListResponse {
  templates: MarketplaceTemplate[]
  total: number
}

// Review list response
export interface ReviewListResponse {
  reviews: TemplateReview[]
  total: number
}
