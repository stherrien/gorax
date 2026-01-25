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
  rating1Count: number
  rating2Count: number
  rating3Count: number
  rating4Count: number
  rating5Count: number
  isVerified: boolean
  isFeatured: boolean
  featuredAt?: string
  featuredBy?: string
  sourceTenantId?: string
  sourceTemplateId?: string
  publishedAt: string
  updatedAt: string
}

// Category type
export interface Category {
  id: string
  name: string
  slug: string
  description: string
  icon: string
  parentId?: string
  displayOrder: number
  templateCount: number
  createdAt: string
  updatedAt: string
  children?: Category[]
}

// Marketplace template with categories
export interface MarketplaceTemplateWithCategories extends MarketplaceTemplate {
  categories: Category[]
}

// Template category type (legacy)
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
  helpfulCount: number
  isHidden: boolean
  hiddenReason?: string
  hiddenAt?: string
  hiddenBy?: string
  deletedAt?: string
  createdAt: string
  updatedAt: string
}

// Review sort options
export type ReviewSortOption = 'recent' | 'helpful' | 'rating_high' | 'rating_low'

// Rating distribution
export interface RatingDistribution {
  rating1Count: number
  rating2Count: number
  rating3Count: number
  rating4Count: number
  rating5Count: number
  totalRatings: number
  averageRating: number
  rating1Percent: number
  rating2Percent: number
  rating3Percent: number
  rating4Percent: number
  rating5Percent: number
}

// Review report
export interface ReviewReport {
  id: string
  reviewId: string
  reporterTenantId: string
  reporterUserId: string
  reason: ReviewReportReason
  details: string
  status: ReviewReportStatus
  resolvedAt?: string
  resolvedBy?: string
  resolutionNotes?: string
  createdAt: string
}

export type ReviewReportReason = 'spam' | 'inappropriate' | 'offensive' | 'misleading' | 'other'
export type ReviewReportStatus = 'pending' | 'reviewed' | 'actioned' | 'dismissed'

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

// Report review input
export interface ReportReviewInput {
  reason: ReviewReportReason
  details?: string
}

// Create category input
export interface CreateCategoryInput {
  name: string
  slug: string
  description?: string
  icon?: string
  parentId?: string
  displayOrder?: number
}

// Update category input
export interface UpdateCategoryInput {
  name?: string
  slug?: string
  description?: string
  icon?: string
  parentId?: string
  displayOrder?: number
}

// Search filter
export interface SearchFilter {
  category?: string
  tags?: string[]
  searchQuery?: string
  minRating?: number
  isVerified?: boolean
  isFeatured?: boolean
  sortBy?: 'popular' | 'recent' | 'rating'
  page?: number
  limit?: number
}

// Enhanced search filter with category IDs
export interface EnhancedSearchFilter {
  categoryIds?: string[]
  tags?: string[]
  searchQuery?: string
  minRating?: number
  isVerified?: boolean
  isFeatured?: boolean
  sortBy?: 'popular' | 'recent' | 'rating' | 'name' | 'relevance'
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
