# Product Requirements Document (PRD)
## Gorax - Workflow Automation Platform

**Version**: 1.0
**Last Updated**: January 12, 2026
**Status**: Living Document
**Owner**: Engineering Team

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Problem Statement](#2-problem-statement)
3. [Product Vision & Goals](#3-product-vision--goals)
4. [Target Users & Personas](#4-target-users--personas)
5. [User Stories & Use Cases](#5-user-stories--use-cases)
6. [Functional Requirements](#6-functional-requirements)
7. [Non-Functional Requirements](#7-non-functional-requirements)
8. [Technical Architecture](#8-technical-architecture)
9. [Feature Specifications](#9-feature-specifications)
10. [API Specification](#10-api-specification)
11. [Security Requirements](#11-security-requirements)
12. [Success Metrics & KPIs](#12-success-metrics--kpis)
13. [Roadmap](#13-roadmap)
14. [Risks & Mitigations](#14-risks--mitigations)
15. [Glossary](#15-glossary)
16. [Appendices](#16-appendices)

---

## 1. Executive Summary

### 1.1 Product Overview

**Gorax** is an open-source, enterprise-grade workflow automation platform that enables organizations to build, deploy, and manage complex workflows through a visual drag-and-drop interface. Designed as a modern alternative to proprietary solutions like Tines, Gorax democratizes workflow automation by providing an accessible, secure, and scalable platform for teams of all technical levels.

### 1.2 Value Proposition

| Stakeholder | Value |
|-------------|-------|
| **DevOps Engineers** | Automate deployment pipelines, incident response, and infrastructure monitoring |
| **IT Operations** | Streamline ticket routing, service health checks, and alert management |
| **Business Analysts** | Build approval workflows and data synchronization without coding |
| **Security Teams** | Automate security incident response and compliance workflows |
| **Platform Teams** | Provide self-service automation capabilities to the organization |

### 1.3 Key Differentiators

1. **Open Source**: Full transparency, community-driven development, no vendor lock-in
2. **No-Code/Low-Code**: Visual builder accessible to non-developers
3. **Enterprise Security**: AES-256-GCM encryption, RBAC, multi-tenancy, audit logging
4. **Cloud-Native**: Kubernetes-ready with horizontal scaling capabilities
5. **Rich Integrations**: 20+ pre-built integrations with major platforms
6. **Real-Time Collaboration**: WebSocket-based live updates and concurrent editing
7. **Self-Hosted Option**: Full control over data and infrastructure

---

## 2. Problem Statement

### 2.1 Current Challenges

Organizations face significant challenges in managing operational workflows:

| Challenge | Impact |
|-----------|--------|
| **Manual Processes** | Teams spend 40%+ of time on repetitive tasks that could be automated |
| **Tool Fragmentation** | Average enterprise uses 900+ applications, requiring custom integrations |
| **Knowledge Silos** | Automation expertise concentrated in few individuals |
| **Compliance Burden** | Manual processes lack audit trails and consistent execution |
| **Slow Incident Response** | MTTR (Mean Time to Resolution) extended by manual escalation |
| **Scalability Issues** | Manual processes don't scale with business growth |

### 2.2 Market Gap

Existing solutions fall short in key areas:

| Solution Type | Limitation |
|---------------|-----------|
| **Proprietary Platforms** (Tines, Workato) | High cost, vendor lock-in, limited customization |
| **Open-Source Tools** (n8n, Airflow) | Complex setup, limited enterprise features, steep learning curve |
| **Custom Scripts** | Difficult to maintain, no visibility, security risks |
| **iPaaS Solutions** | Limited workflow complexity, expensive at scale |

### 2.3 Opportunity

The workflow automation market is projected to reach $26B by 2028 (CAGR 24.4%). Organizations need a solution that combines:

- Enterprise-grade security and compliance
- Intuitive visual interface for non-developers
- Extensibility for technical users
- Self-hosted deployment options
- Open-source transparency

---

## 3. Product Vision & Goals

### 3.1 Vision Statement

> **"Empower every team to automate their workflows, regardless of technical expertise, with an open platform that prioritizes security, scalability, and user experience."**

### 3.2 Mission

Democratize workflow automation by providing:

1. **Accessible Tools**: No-code interface for business users
2. **Powerful Capabilities**: Advanced features for technical users
3. **Enterprise Security**: SOC2-compliant security controls
4. **Open Development**: Community-driven feature development
5. **Operational Excellence**: Built-in observability and reliability

### 3.3 Strategic Goals

| Goal | Metric | Target |
|------|--------|--------|
| **User Adoption** | Monthly Active Users | 10,000 MAU (Year 1) |
| **Platform Reliability** | Uptime | 99.9% availability |
| **Execution Performance** | P95 Latency | < 500ms per action |
| **Community Growth** | GitHub Stars | 5,000 stars (Year 1) |
| **Enterprise Adoption** | Enterprise Customers | 50 paid deployments (Year 1) |

### 3.4 Success Criteria

**For Users:**
- Create first workflow in < 15 minutes
- Reduce manual tasks by 50%+ within 90 days
- Achieve 95%+ workflow execution success rate

**For Organizations:**
- Decrease MTTR by 40% through automated incident response
- Reduce integration development time by 70%
- Achieve full audit compliance for automated processes

---

## 4. Target Users & Personas

### 4.1 Primary Personas

#### Persona 1: DevOps Engineer ("Alex")

| Attribute | Description |
|-----------|-------------|
| **Role** | Senior DevOps Engineer |
| **Team Size** | 5-15 engineers |
| **Technical Level** | Advanced |
| **Goals** | Automate CI/CD, reduce toil, improve incident response |
| **Pain Points** | Too many manual deployment tasks, inconsistent processes, alert fatigue |
| **Success Metrics** | Deployment frequency, MTTR, automation coverage |

**Key Workflows:**
- Deployment notifications to Slack
- Automated rollback on failure detection
- Infrastructure health check orchestration
- On-call rotation management

#### Persona 2: IT Operations Manager ("Jordan")

| Attribute | Description |
|-----------|-------------|
| **Role** | IT Operations Manager |
| **Team Size** | 10-30 staff |
| **Technical Level** | Intermediate |
| **Goals** | Streamline service desk, reduce ticket backlog, improve SLAs |
| **Pain Points** | Manual ticket routing, slow escalation, lack of visibility |
| **Success Metrics** | Ticket resolution time, SLA compliance, team productivity |

**Key Workflows:**
- Automated ticket routing and prioritization
- Service health monitoring and alerting
- Approval workflows for access requests
- Report generation and distribution

#### Persona 3: Business Analyst ("Sam")

| Attribute | Description |
|-----------|-------------|
| **Role** | Business Process Analyst |
| **Team Size** | 3-10 analysts |
| **Technical Level** | Beginner |
| **Goals** | Automate business processes without IT dependency |
| **Pain Points** | Waiting on IT for automation, complex approval chains, data sync issues |
| **Success Metrics** | Process cycle time, error reduction, automation adoption |

**Key Workflows:**
- Customer onboarding approval flows
- Data synchronization between systems
- Report scheduling and distribution
- Compliance audit workflows

#### Persona 4: Security Operations ("Casey")

| Attribute | Description |
|-----------|-------------|
| **Role** | Security Operations Analyst |
| **Team Size** | 3-10 analysts |
| **Technical Level** | Advanced |
| **Goals** | Automate security incident response, reduce dwell time |
| **Pain Points** | Manual threat investigation, slow containment, alert overload |
| **Success Metrics** | MTTD, MTTR, false positive rate, incidents automated |

**Key Workflows:**
- SIEM alert enrichment and triage
- Automated IOC investigation
- Incident escalation and notification
- Compliance evidence collection

### 4.2 Secondary Personas

| Persona | Use Case | Technical Level |
|---------|----------|-----------------|
| **Platform Engineer** | Building internal automation tools | Advanced |
| **Customer Support Lead** | Ticket automation and escalation | Beginner |
| **Compliance Officer** | Audit workflow management | Intermediate |
| **Data Engineer** | ETL pipeline orchestration | Advanced |

### 4.3 Anti-Personas (Not Target Users)

| Profile | Reason |
|---------|--------|
| **Developers needing full code control** | Better served by Airflow, Prefect |
| **Simple task schedulers** | Cron or CloudWatch Events sufficient |
| **Real-time stream processing** | Kafka Streams, Flink more appropriate |

---

## 5. User Stories & Use Cases

### 5.1 Core User Stories

#### Workflow Creation

| ID | Story | Priority |
|----|-------|----------|
| US-001 | As a DevOps engineer, I want to create workflows visually so that I can automate processes without writing code | P0 |
| US-002 | As a business analyst, I want to drag-and-drop actions so that I can build workflows quickly | P0 |
| US-003 | As a power user, I want to write custom JavaScript actions so that I can handle complex logic | P1 |
| US-004 | As a team lead, I want to use templates so that my team can start from proven patterns | P1 |

#### Workflow Execution

| ID | Story | Priority |
|----|-------|----------|
| US-010 | As an operator, I want to trigger workflows via webhooks so that external systems can initiate automation | P0 |
| US-011 | As a scheduler, I want to run workflows on a cron schedule so that reports generate automatically | P0 |
| US-012 | As a manager, I want approval steps in workflows so that human oversight is preserved | P1 |
| US-013 | As an operator, I want to monitor workflow execution in real-time so that I can identify issues quickly | P0 |

#### Integration Management

| ID | Story | Priority |
|----|-------|----------|
| US-020 | As an admin, I want to securely store API credentials so that workflows can authenticate safely | P0 |
| US-021 | As a user, I want pre-built integrations so that I don't have to configure HTTP requests manually | P0 |
| US-022 | As a developer, I want to create custom integrations so that I can connect proprietary systems | P1 |

#### Administration

| ID | Story | Priority |
|----|-------|----------|
| US-030 | As an admin, I want role-based access control so that users only access appropriate resources | P0 |
| US-031 | As an auditor, I want execution history so that I can review what actions were taken | P0 |
| US-032 | As a compliance officer, I want audit logs so that I can demonstrate regulatory compliance | P0 |

### 5.2 Use Case Scenarios

#### Use Case 1: Deployment Notification Workflow

**Actor:** DevOps Engineer (Alex)

**Preconditions:**
- User has Slack integration configured
- User has GitHub webhook set up

**Flow:**
1. GitHub sends webhook on deployment event
2. Workflow extracts deployment details (repo, environment, status)
3. Condition node checks deployment status
4. If successful: Send success message to #deployments channel
5. If failed: Send alert to #incidents channel with PagerDuty page

**Postconditions:**
- Team notified of deployment status
- Failed deployments trigger incident response

**Metrics:**
- Notification latency < 5 seconds
- 100% webhook processing reliability

#### Use Case 2: Automated Ticket Routing

**Actor:** IT Operations Manager (Jordan)

**Preconditions:**
- ServiceNow integration configured
- Routing rules defined

**Flow:**
1. New ticket created in ServiceNow triggers workflow
2. AI action analyzes ticket content
3. Category and priority assigned automatically
4. Ticket routed to appropriate team queue
5. SLA timer started
6. Notification sent to assigned team

**Postconditions:**
- Ticket categorized and routed within 60 seconds
- Team notified immediately

**Metrics:**
- Routing accuracy > 95%
- Mean routing time < 30 seconds

#### Use Case 3: Security Incident Response

**Actor:** Security Operations (Casey)

**Preconditions:**
- SIEM webhook configured
- Threat intel feeds connected

**Flow:**
1. SIEM sends high-severity alert
2. Workflow enriches IOCs from threat intel
3. Automated containment action (firewall block)
4. Incident created in ITSM
5. On-call security analyst paged
6. Evidence collected and attached to incident

**Postconditions:**
- Threat contained within 5 minutes
- Full incident documentation created

**Metrics:**
- MTTR reduced by 60%
- Containment time < 5 minutes

#### Use Case 4: Approval Workflow with Human Task

**Actor:** Business Analyst (Sam)

**Preconditions:**
- Approval workflow configured
- Approvers assigned

**Flow:**
1. Request submitted via API/form
2. Workflow sends approval request to manager
3. Manager receives notification with approve/reject options
4. If approved: Proceed to next step
5. If rejected: Notify requester with reason
6. If timeout (48h): Escalate to director

**Postconditions:**
- Request processed with full audit trail
- Escalation ensures no requests stuck

**Metrics:**
- Approval completion rate > 99%
- Mean approval time < 24 hours

---

## 6. Functional Requirements

### 6.1 Workflow Builder

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-001 | Visual drag-and-drop canvas for workflow design | P0 | Complete |
| FR-002 | Support for trigger, action, condition, loop, and parallel nodes | P0 | Complete |
| FR-003 | Real-time validation of workflow structure (DAG validation) | P0 | Complete |
| FR-004 | Expression language for dynamic values (CEL) | P0 | Complete |
| FR-005 | Node configuration panel with form validation | P0 | Complete |
| FR-006 | Undo/redo support for canvas operations | P1 | Complete |
| FR-007 | Workflow versioning with history | P1 | Complete |
| FR-008 | Import/export workflows as JSON | P1 | Complete |
| FR-009 | Workflow templates and marketplace | P2 | Complete |
| FR-010 | Collaborative editing with WebSocket sync | P2 | Complete |

### 6.2 Workflow Execution

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-020 | Synchronous workflow execution with timeout | P0 | Complete |
| FR-021 | Asynchronous execution via job queue | P0 | Complete |
| FR-022 | Conditional branching (if/else) | P0 | Complete |
| FR-023 | Loop execution (for each, while) | P0 | Complete |
| FR-024 | Parallel execution with join semantics | P1 | Complete |
| FR-025 | Sub-workflow invocation with depth limiting | P1 | Complete |
| FR-026 | Error handling (try/catch/finally) | P1 | Complete |
| FR-027 | Retry logic with exponential backoff | P1 | Complete |
| FR-028 | Circuit breaker for failing integrations | P2 | Complete |
| FR-029 | Human task/approval steps | P1 | Complete |
| FR-030 | Execution cancellation | P1 | Complete |

### 6.3 Triggers

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-040 | Webhook triggers with signature validation | P0 | Complete |
| FR-041 | Scheduled triggers (cron expressions) | P0 | Complete |
| FR-042 | Manual trigger via API | P0 | Complete |
| FR-043 | Event filtering with JSONPath | P1 | Complete |
| FR-044 | Webhook replay functionality | P2 | Complete |
| FR-045 | Webhook test/simulation | P1 | Complete |

### 6.4 Integrations

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-060 | HTTP/REST action (generic API calls) | P0 | Complete |
| FR-061 | Slack integration (messages, reactions) | P0 | Complete |
| FR-062 | GitHub integration (issues, PRs) | P0 | Complete |
| FR-063 | Jira integration (issues, transitions) | P0 | Complete |
| FR-064 | PagerDuty integration (incidents) | P1 | Complete |
| FR-065 | Email providers (SendGrid, SES, SMTP) | P0 | Complete |
| FR-066 | SMS providers (Twilio, SNS) | P1 | Complete |
| FR-067 | AWS services (Lambda, SQS, S3) | P1 | Complete |
| FR-068 | JavaScript code execution (sandbox) | P1 | Complete |
| FR-069 | AI/LLM actions (OpenAI, Anthropic) | P2 | Complete |
| FR-070 | Database actions (query, insert) | P2 | Complete |

### 6.5 Credential Management

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-080 | Encrypted credential storage (AES-256-GCM) | P0 | Complete |
| FR-081 | Credential injection via template syntax | P0 | Complete |
| FR-082 | Credential masking in logs | P0 | Complete |
| FR-083 | AWS KMS integration for key management | P1 | Complete |
| FR-084 | Credential rotation support | P2 | Planned |
| FR-085 | OAuth 2.0 token management | P2 | Planned |

### 6.6 Monitoring & Observability

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-100 | Real-time execution status via WebSocket | P0 | Complete |
| FR-101 | Execution history with step-level detail | P0 | Complete |
| FR-102 | Execution metrics dashboard | P1 | Complete |
| FR-103 | Prometheus metrics endpoint | P1 | Complete |
| FR-104 | OpenTelemetry distributed tracing | P2 | Complete |
| FR-105 | Error tracking (Sentry integration) | P2 | Complete |
| FR-106 | Alerting on workflow failures | P1 | Complete |

### 6.7 Administration

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-120 | Multi-tenancy with data isolation | P0 | Complete |
| FR-121 | Role-based access control (RBAC) | P0 | Complete |
| FR-122 | Audit logging for all actions | P0 | Complete |
| FR-123 | User management (Ory Kratos) | P0 | Complete |
| FR-124 | Retention policies for execution data | P1 | Complete |
| FR-125 | Usage quotas and rate limiting | P1 | Complete |

---

## 7. Non-Functional Requirements

### 7.1 Performance

| Metric | Requirement | Current |
|--------|-------------|---------|
| **Webhook Response Time** | P95 < 200ms | 150ms |
| **Action Execution Time** | P95 < 500ms (excluding external calls) | 380ms |
| **Workflow Start Latency** | P95 < 100ms | 75ms |
| **Concurrent Executions** | 10,000+ per tenant | 15,000 |
| **Dashboard Load Time** | < 2 seconds | 1.2s |
| **Canvas Render Time** | < 500ms for 100-node workflow | 320ms |

### 7.2 Scalability

| Dimension | Requirement |
|-----------|-------------|
| **Horizontal Scaling** | Stateless API servers behind load balancer |
| **Database Scaling** | Read replicas for query distribution |
| **Queue Scaling** | SQS auto-scaling based on queue depth |
| **Storage Scaling** | S3/MinIO for unlimited object storage |
| **Max Workflows per Tenant** | 10,000 |
| **Max Executions per Day** | 1,000,000 per tenant |

### 7.3 Availability

| Metric | Requirement |
|--------|-------------|
| **Uptime SLA** | 99.9% (8.76 hours downtime/year) |
| **RTO (Recovery Time)** | < 15 minutes |
| **RPO (Recovery Point)** | < 5 minutes |
| **Failover** | Automatic database failover |
| **Backup Frequency** | Continuous WAL archiving |

### 7.4 Security

| Requirement | Implementation |
|-------------|----------------|
| **Encryption at Rest** | AES-256-GCM for credentials, PostgreSQL TDE |
| **Encryption in Transit** | TLS 1.3 for all connections |
| **Authentication** | Ory Kratos with MFA support |
| **Authorization** | RBAC with tenant isolation |
| **Secrets Management** | Envelope encryption with AWS KMS |
| **Vulnerability Scanning** | Trivy, Dependabot, CodeQL |
| **Penetration Testing** | Annual third-party assessment |

### 7.5 Compliance

| Standard | Status |
|----------|--------|
| **SOC 2 Type II** | Planned |
| **GDPR** | Compliant (data residency, right to deletion) |
| **HIPAA** | Compliant (with BAA) |
| **ISO 27001** | Planned |

### 7.6 Reliability

| Mechanism | Implementation |
|-----------|----------------|
| **Circuit Breaker** | Per-integration failure detection |
| **Retry Logic** | Exponential backoff with jitter |
| **Dead Letter Queue** | Failed executions preserved for retry |
| **Graceful Degradation** | Feature flags for integration failures |
| **Health Checks** | Kubernetes liveness/readiness probes |

### 7.7 Usability

| Metric | Target |
|--------|--------|
| **Time to First Workflow** | < 15 minutes |
| **Documentation Coverage** | 100% of features |
| **Error Message Clarity** | Actionable with suggested fixes |
| **Accessibility** | WCAG 2.1 AA compliance |

---

## 8. Technical Architecture

### 8.1 System Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              PRESENTATION LAYER                              │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │                        React Frontend (Vite)                          │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │  │
│  │  │   Canvas    │  │  Dashboard  │  │  Settings   │  │   Monitor   │  │  │
│  │  │  (ReactFlow)│  │             │  │             │  │             │  │  │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘  │  │
│  │                                                                       │  │
│  │  ┌─────────────────────────────────────────────────────────────────┐ │  │
│  │  │  State: Zustand  │  Server State: TanStack Query  │  WS: Live   │ │  │
│  │  └─────────────────────────────────────────────────────────────────┘ │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                                 API LAYER                                    │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │                     Chi Router (REST API v1)                          │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────┐  │  │
│  │  │ Workflow │  │ Webhook  │  │Execution │  │Credential│  │ Health │  │  │
│  │  │ Handlers │  │ Handlers │  │ Handlers │  │ Handlers │  │  Check │  │  │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘  └────────┘  │  │
│  │                                                                       │  │
│  │  ┌─────────────────────────────────────────────────────────────────┐ │  │
│  │  │ Middleware: Auth │ Tenant │ RBAC │ Rate Limit │ Security │ CORS │ │  │
│  │  └─────────────────────────────────────────────────────────────────┘ │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              SERVICE LAYER                                   │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │  Workflow   │  │  Execution  │  │  Credential │  │     Integration     │ │
│  │   Service   │  │   Engine    │  │   Service   │  │      Registry       │ │
│  │             │  │   (Visitor) │  │  (Encrypt)  │  │                     │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────────────┘ │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │   Webhook   │  │  Schedule   │  │   RBAC      │  │      Analytics      │ │
│  │   Service   │  │   Service   │  │   Service   │  │       Service       │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                             DOMAIN LAYER                                     │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │  Workflow   │  │   Node      │  │  Execution  │  │      Tenant         │ │
│  │   Entity    │  │   Types     │  │   Entity    │  │       Entity        │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────────────┘ │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │ Credential  │  │   Webhook   │  │  Schedule   │  │      User           │ │
│  │   Entity    │  │   Entity    │  │   Entity    │  │      Entity         │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           INFRASTRUCTURE LAYER                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │ PostgreSQL  │  │    Redis    │  │   AWS SQS   │  │      S3/MinIO       │ │
│  │  Repository │  │    Cache    │  │    Queue    │  │       Storage       │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────────────┘ │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │ Ory Kratos  │  │  AWS KMS    │  │  Prometheus │  │       Jaeger        │ │
│  │    Auth     │  │    Keys     │  │   Metrics   │  │      Tracing        │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 8.2 Technology Stack

#### Backend

| Component | Technology | Version |
|-----------|------------|---------|
| Language | Go | 1.24 |
| HTTP Router | Chi | v5 |
| Database | PostgreSQL | 16 |
| Cache | Redis | 7 |
| Queue | AWS SQS | - |
| Auth | Ory Kratos | Latest |
| Validation | go-playground/validator | v10 |
| Logging | slog (stdlib) | - |
| GraphQL | gqlgen | - |

#### Frontend

| Component | Technology | Version |
|-----------|------------|---------|
| Framework | React | 18 |
| Language | TypeScript | 5 |
| Build Tool | Vite | 7 |
| State Management | Zustand | Latest |
| Server State | TanStack Query | v5 |
| Canvas | ReactFlow | v12 |
| Styling | Tailwind CSS | 3 |
| Forms | React Hook Form + Zod | - |
| Testing | Vitest + Testing Library | - |

#### Infrastructure

| Component | Technology |
|-----------|------------|
| Container | Docker |
| Orchestration | Kubernetes |
| Storage | S3/MinIO |
| Monitoring | Prometheus + Grafana |
| Tracing | Jaeger (OpenTelemetry) |
| Error Tracking | Sentry |
| CI/CD | GitHub Actions |

### 8.3 Data Model

```
┌─────────────────────┐     ┌─────────────────────┐
│       Tenant        │     │        User         │
├─────────────────────┤     ├─────────────────────┤
│ id (PK)             │◄────│ tenant_id (FK)      │
│ name                │     │ id (PK)             │
│ settings (JSONB)    │     │ email               │
│ created_at          │     │ role                │
│ updated_at          │     │ created_at          │
└─────────────────────┘     └─────────────────────┘
          │                           │
          │                           │
          ▼                           ▼
┌─────────────────────┐     ┌─────────────────────┐
│      Workflow       │     │     Credential      │
├─────────────────────┤     ├─────────────────────┤
│ id (PK)             │     │ id (PK)             │
│ tenant_id (FK)      │     │ tenant_id (FK)      │
│ name                │     │ name                │
│ description         │     │ type                │
│ graph (JSONB)       │     │ encrypted_value     │
│ status              │     │ created_by (FK)     │
│ version             │     │ created_at          │
│ created_by (FK)     │     └─────────────────────┘
│ created_at          │
│ updated_at          │
└─────────────────────┘
          │
          │
          ▼
┌─────────────────────┐     ┌─────────────────────┐
│      Execution      │────►│   ExecutionStep     │
├─────────────────────┤     ├─────────────────────┤
│ id (PK)             │     │ id (PK)             │
│ workflow_id (FK)    │     │ execution_id (FK)   │
│ tenant_id (FK)      │     │ node_id             │
│ status              │     │ status              │
│ trigger_type        │     │ input (JSONB)       │
│ input (JSONB)       │     │ output (JSONB)      │
│ output (JSONB)      │     │ error               │
│ error               │     │ started_at          │
│ started_at          │     │ completed_at        │
│ completed_at        │     └─────────────────────┘
└─────────────────────┘
          │
          │
          ▼
┌─────────────────────┐     ┌─────────────────────┐
│       Webhook       │     │      Schedule       │
├─────────────────────┤     ├─────────────────────┤
│ id (PK)             │     │ id (PK)             │
│ workflow_id (FK)    │     │ workflow_id (FK)    │
│ tenant_id (FK)      │     │ tenant_id (FK)      │
│ endpoint            │     │ cron_expression     │
│ secret              │     │ timezone            │
│ filter (JSONB)      │     │ enabled             │
│ enabled             │     │ next_run_at         │
│ created_at          │     │ last_run_at         │
└─────────────────────┘     └─────────────────────┘
```

### 8.4 Execution Engine

The workflow execution engine uses the **Visitor Pattern** for node traversal:

```go
// Node interface
type Node interface {
    Accept(visitor NodeVisitor) error
}

// Visitor interface
type NodeVisitor interface {
    VisitTrigger(node *TriggerNode) error
    VisitAction(node *ActionNode) error
    VisitCondition(node *ConditionNode) error
    VisitLoop(node *LoopNode) error
    VisitParallel(node *ParallelNode) error
}

// ExecutionVisitor implements the visitor
type ExecutionVisitor struct {
    context    *ExecutionContext
    registry   ActionRegistry
    credential CredentialService
}
```

**Execution Flow:**
1. Parse workflow graph from JSON
2. Validate DAG structure (no cycles)
3. Initialize execution context
4. Execute nodes in topological order
5. Handle branching and parallel paths
6. Record step results
7. Handle errors with retry/circuit breaker

---

## 9. Feature Specifications

### 9.1 Visual Workflow Builder

#### 9.1.1 Canvas Interface

**Description:** A drag-and-drop canvas for designing workflows using ReactFlow.

**Capabilities:**
- Node palette with categorized actions
- Drag-and-drop node placement
- Edge connections with validation
- Mini-map for navigation
- Zoom and pan controls
- Grid snapping
- Node grouping

**Node Types:**

| Type | Icon | Description |
|------|------|-------------|
| Trigger | ⚡ | Entry point (webhook, schedule, manual) |
| Action | ▶️ | Execute integration (HTTP, Slack, etc.) |
| Condition | ◆ | Branch based on expression |
| Loop | 🔄 | Iterate over collection |
| Parallel | ⫸ | Execute branches concurrently |
| Human Task | 👤 | Wait for manual approval |
| Sub-Workflow | 📋 | Invoke another workflow |

#### 9.1.2 Node Configuration

**Panel Components:**
- Node name and description
- Integration type selector
- Input parameter form (dynamic based on action)
- Output mapping configuration
- Error handling settings
- Retry configuration

**Expression Language (CEL):**
```
{{trigger.payload.user.email}}
{{steps.http_request.response.body.id}}
{{env.SLACK_CHANNEL}}
{{credentials.github_token}}
```

#### 9.1.3 Workflow Validation

**Real-time Validation:**
- DAG cycle detection
- Required field validation
- Expression syntax validation
- Connection type compatibility
- Credential reference validation

### 9.2 Integrations

#### 9.2.1 Communication Integrations

**Slack**
| Action | Parameters | Output |
|--------|------------|--------|
| send_message | channel, text, blocks | message_ts, channel |
| send_dm | user_id, text | message_ts |
| add_reaction | channel, timestamp, emoji | ok |
| update_message | channel, ts, text | ok |

**Email (SendGrid/SES/SMTP)**
| Action | Parameters | Output |
|--------|------------|--------|
| send_email | to, subject, body, html | message_id |
| send_template | template_id, to, variables | message_id |

**SMS (Twilio/SNS)**
| Action | Parameters | Output |
|--------|------------|--------|
| send_sms | to, message | sid |

#### 9.2.2 Issue Tracking Integrations

**GitHub**
| Action | Parameters | Output |
|--------|------------|--------|
| create_issue | owner, repo, title, body | issue_number, url |
| add_label | owner, repo, issue, label | ok |
| create_comment | owner, repo, issue, body | comment_id |

**Jira**
| Action | Parameters | Output |
|--------|------------|--------|
| create_issue | project, type, summary, description | key, id |
| transition_issue | issue_key, transition_id | ok |
| add_comment | issue_key, body | comment_id |
| search | jql, max_results | issues[] |

**PagerDuty**
| Action | Parameters | Output |
|--------|------------|--------|
| create_incident | service_id, title, urgency | incident_id |
| acknowledge | incident_id | ok |
| resolve | incident_id | ok |

#### 9.2.3 Cloud Integrations

**AWS**
| Action | Parameters | Output |
|--------|------------|--------|
| invoke_lambda | function_name, payload | response |
| publish_sns | topic_arn, message | message_id |
| send_sqs | queue_url, message | message_id |
| s3_get_object | bucket, key | body |
| s3_put_object | bucket, key, body | etag |

#### 9.2.4 AI/LLM Integrations

**OpenAI/Anthropic/Bedrock**
| Action | Parameters | Output |
|--------|------------|--------|
| chat_completion | model, messages, temperature | content, tokens_used |
| embedding | model, input | embedding[] |
| entity_extraction | model, text, schema | entities |

### 9.3 Human Tasks / Approvals

**Configuration:**
```yaml
type: human_task
config:
  assignees: ["user@example.com"]
  timeout: 48h
  escalation:
    - after: 24h
      notify: ["manager@example.com"]
    - after: 48h
      action: auto_reject
  options:
    - label: Approve
      value: approved
      style: success
    - label: Reject
      value: rejected
      style: danger
```

**Notification Channels:**
- Email with action buttons
- Slack with interactive blocks
- Mobile push notification
- In-app notification center

**Audit Trail:**
- Who approved/rejected
- When action was taken
- Comments provided
- Full payload at decision time

### 9.4 Marketplace

**Template Categories:**
- DevOps & CI/CD
- Incident Management
- Business Processes
- Security Operations
- Customer Support
- Data Integration

**Template Structure:**
```yaml
name: deployment-notifier
version: 1.0.0
description: Notify team of deployments via Slack
author: gorax-team
category: devops
tags: [deployment, slack, github]
required_credentials:
  - type: slack
    name: SLACK_BOT_TOKEN
  - type: github
    name: GITHUB_WEBHOOK_SECRET
workflow:
  # ... workflow JSON
```

---

## 10. API Specification

### 10.1 API Overview

| Property | Value |
|----------|-------|
| Base URL | `/api/v1` |
| Authentication | Ory Kratos session + `X-Tenant-ID` header |
| Rate Limiting | 1000 requests/minute per tenant |
| Content Type | `application/json` |

### 10.2 Endpoints Summary

#### Workflows
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/workflows` | List workflows |
| POST | `/workflows` | Create workflow |
| GET | `/workflows/{id}` | Get workflow |
| PUT | `/workflows/{id}` | Update workflow |
| DELETE | `/workflows/{id}` | Delete workflow |
| POST | `/workflows/{id}/execute` | Execute workflow |
| GET | `/workflows/{id}/versions` | List versions |
| POST | `/workflows/{id}/clone` | Clone workflow |

#### Executions
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/executions` | List executions |
| GET | `/executions/{id}` | Get execution details |
| POST | `/executions/{id}/cancel` | Cancel execution |
| POST | `/executions/{id}/retry` | Retry failed execution |
| GET | `/executions/{id}/steps` | Get step history |

#### Webhooks
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/webhooks` | List webhooks |
| POST | `/webhooks` | Create webhook |
| GET | `/webhooks/{id}` | Get webhook |
| PUT | `/webhooks/{id}` | Update webhook |
| DELETE | `/webhooks/{id}` | Delete webhook |
| POST | `/webhooks/{id}/test` | Test webhook |
| GET | `/webhooks/{id}/events` | Event history |
| POST | `/webhook/{endpoint}` | Receive webhook |

#### Schedules
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/schedules` | List schedules |
| POST | `/schedules` | Create schedule |
| GET | `/schedules/{id}` | Get schedule |
| PUT | `/schedules/{id}` | Update schedule |
| DELETE | `/schedules/{id}` | Delete schedule |
| POST | `/schedules/{id}/trigger` | Trigger immediately |

#### Credentials
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/credentials` | List credentials (names only) |
| POST | `/credentials` | Store credential |
| PUT | `/credentials/{id}` | Update credential |
| DELETE | `/credentials/{id}` | Delete credential |

#### Analytics
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/analytics/executions` | Execution metrics |
| GET | `/analytics/workflows` | Workflow usage stats |
| GET | `/analytics/integrations` | Integration usage |

### 10.3 Request/Response Examples

#### Create Workflow

**Request:**
```http
POST /api/v1/workflows
Content-Type: application/json
X-Tenant-ID: tenant_123

{
  "name": "Deployment Notifier",
  "description": "Notify team of deployments",
  "graph": {
    "nodes": [
      {
        "id": "trigger_1",
        "type": "webhook",
        "position": {"x": 100, "y": 100},
        "config": {
          "method": "POST",
          "path": "/deploy"
        }
      },
      {
        "id": "action_1",
        "type": "slack",
        "position": {"x": 100, "y": 250},
        "config": {
          "action": "send_message",
          "channel": "#deployments",
          "text": "Deployment: {{trigger.payload.repo}} to {{trigger.payload.env}}"
        }
      }
    ],
    "edges": [
      {"source": "trigger_1", "target": "action_1"}
    ]
  }
}
```

**Response:**
```json
{
  "id": "wf_abc123",
  "name": "Deployment Notifier",
  "description": "Notify team of deployments",
  "status": "active",
  "version": 1,
  "created_at": "2026-01-12T10:00:00Z",
  "updated_at": "2026-01-12T10:00:00Z",
  "webhook_url": "https://api.gorax.io/webhook/wh_xyz789"
}
```

#### Execute Workflow

**Request:**
```http
POST /api/v1/workflows/wf_abc123/execute
Content-Type: application/json

{
  "input": {
    "repo": "gorax/gorax",
    "env": "production",
    "commit": "abc123"
  },
  "async": false
}
```

**Response:**
```json
{
  "execution_id": "exec_def456",
  "status": "completed",
  "started_at": "2026-01-12T10:05:00Z",
  "completed_at": "2026-01-12T10:05:02Z",
  "output": {
    "slack_message_ts": "1234567890.123456"
  },
  "steps": [
    {
      "node_id": "trigger_1",
      "status": "completed",
      "duration_ms": 5
    },
    {
      "node_id": "action_1",
      "status": "completed",
      "duration_ms": 1500
    }
  ]
}
```

### 10.4 Error Responses

```json
{
  "error": {
    "code": "WORKFLOW_NOT_FOUND",
    "message": "Workflow with ID 'wf_invalid' not found",
    "details": {
      "workflow_id": "wf_invalid"
    },
    "request_id": "req_abc123"
  }
}
```

| Code | HTTP Status | Description |
|------|-------------|-------------|
| UNAUTHORIZED | 401 | Missing or invalid authentication |
| FORBIDDEN | 403 | Insufficient permissions |
| NOT_FOUND | 404 | Resource not found |
| VALIDATION_ERROR | 400 | Invalid request parameters |
| RATE_LIMITED | 429 | Rate limit exceeded |
| INTERNAL_ERROR | 500 | Server error |

---

## 11. Security Requirements

### 11.1 Authentication

| Mechanism | Description |
|-----------|-------------|
| **Ory Kratos** | Production authentication with MFA |
| **Session Management** | Secure httpOnly cookies |
| **Token Expiry** | 24-hour session validity |
| **MFA** | TOTP/WebAuthn support |

### 11.2 Authorization

**RBAC Roles:**
| Role | Permissions |
|------|-------------|
| Admin | Full access, user management, settings |
| Editor | Create/edit/delete workflows, view executions |
| Viewer | Read-only access to workflows and executions |
| Operator | Execute workflows, view executions |

**Permission Matrix:**
| Resource | Admin | Editor | Viewer | Operator |
|----------|-------|--------|--------|----------|
| Workflows (Create) | ✅ | ✅ | ❌ | ❌ |
| Workflows (Read) | ✅ | ✅ | ✅ | ✅ |
| Workflows (Update) | ✅ | ✅ | ❌ | ❌ |
| Workflows (Delete) | ✅ | ✅ | ❌ | ❌ |
| Executions (Execute) | ✅ | ✅ | ❌ | ✅ |
| Executions (View) | ✅ | ✅ | ✅ | ✅ |
| Credentials (Manage) | ✅ | ✅ | ❌ | ❌ |
| Users (Manage) | ✅ | ❌ | ❌ | ❌ |

### 11.3 Data Protection

**Encryption:**
| Data | Encryption |
|------|------------|
| Credentials | AES-256-GCM (envelope encryption) |
| Database | PostgreSQL TDE |
| Transit | TLS 1.3 |
| Backups | AES-256 |

**Key Management:**
- AWS KMS for production KEK
- 90-day key rotation
- Separate keys per tenant

### 11.4 Security Controls

| Control | Implementation |
|---------|----------------|
| **Input Validation** | go-playground/validator, Zod |
| **SQL Injection** | Parameterized queries only |
| **XSS** | React auto-escaping, CSP headers |
| **CSRF** | SameSite cookies, CSRF tokens |
| **Rate Limiting** | Per-tenant, per-endpoint limits |
| **Security Headers** | HSTS, X-Frame-Options, CSP |

### 11.5 Audit Logging

**Logged Events:**
- Authentication (login, logout, MFA)
- Authorization (permission checks)
- Workflow changes (create, update, delete)
- Execution events (start, complete, fail)
- Credential access
- Admin actions

**Log Format:**
```json
{
  "timestamp": "2026-01-12T10:00:00Z",
  "event": "workflow.updated",
  "tenant_id": "tenant_123",
  "user_id": "user_456",
  "resource_id": "wf_abc123",
  "ip_address": "192.168.1.1",
  "user_agent": "Mozilla/5.0...",
  "changes": {
    "name": {"old": "Old Name", "new": "New Name"}
  }
}
```

---

## 12. Success Metrics & KPIs

### 12.1 Product Metrics

| Metric | Definition | Target | Measurement |
|--------|------------|--------|-------------|
| **MAU** | Monthly Active Users | 10,000 (Y1) | Unique users/month |
| **WAU/MAU** | Stickiness | > 40% | Weekly/Monthly ratio |
| **Workflows Created** | New workflows | 50,000 (Y1) | Count per period |
| **Execution Volume** | Workflow runs | 10M/month (Y1) | Total executions |
| **Time to First Workflow** | Onboarding speed | < 15 min | Median time |

### 12.2 Engagement Metrics

| Metric | Definition | Target |
|--------|------------|--------|
| **D7 Retention** | Users returning after 7 days | > 50% |
| **D30 Retention** | Users returning after 30 days | > 30% |
| **Workflows per User** | Active workflows per user | > 5 |
| **Executions per Workflow** | Average runs per workflow | > 100/month |

### 12.3 Reliability Metrics

| Metric | Definition | Target |
|--------|------------|--------|
| **Uptime** | Service availability | 99.9% |
| **Execution Success Rate** | Successful completions | > 99% |
| **P95 Latency (Webhook)** | Webhook response time | < 200ms |
| **P95 Latency (Execution)** | Workflow start latency | < 100ms |
| **MTTR** | Mean time to recovery | < 15 min |

### 12.4 Business Metrics

| Metric | Definition | Target (Y1) |
|--------|------------|-------------|
| **Enterprise Customers** | Paid deployments | 50 |
| **ARR** | Annual Recurring Revenue | $500K |
| **NPS** | Net Promoter Score | > 40 |
| **GitHub Stars** | Community interest | 5,000 |
| **Contributors** | Active contributors | 50 |

### 12.5 Tracking Implementation

**Analytics Tools:**
- Product: PostHog for user analytics
- Infrastructure: Prometheus + Grafana
- Error: Sentry
- Logs: Loki

**Dashboard Components:**
- Real-time execution volume
- Success/failure rates by workflow
- Integration usage breakdown
- User activity heatmaps
- Performance percentiles

---

## 13. Roadmap

### 13.1 Phase 1: Foundation (Completed)

**Delivered:**
- ✅ Visual workflow builder with ReactFlow
- ✅ Basic triggers (webhook, schedule, manual)
- ✅ Core actions (HTTP, Slack, Email)
- ✅ Expression language (CEL)
- ✅ Execution engine with visitor pattern
- ✅ PostgreSQL persistence
- ✅ Basic authentication

### 13.2 Phase 2: Enterprise Essentials (Completed)

**Delivered:**
- ✅ Multi-tenancy with data isolation
- ✅ RBAC with fine-grained permissions
- ✅ Encrypted credential management
- ✅ Audit logging
- ✅ Execution history and analytics
- ✅ Additional integrations (GitHub, Jira, PagerDuty)

### 13.3 Phase 3: Advanced Workflows (Completed)

**Delivered:**
- ✅ Conditional branching
- ✅ Loop execution
- ✅ Parallel execution
- ✅ Sub-workflows
- ✅ Error handling (try/catch)
- ✅ Retry with exponential backoff
- ✅ Human tasks/approvals

### 13.4 Phase 4: Scale & Observability (Completed)

**Delivered:**
- ✅ Real-time WebSocket updates
- ✅ Prometheus metrics
- ✅ OpenTelemetry tracing
- ✅ Sentry error tracking
- ✅ Grafana dashboards
- ✅ SQS-based job queue
- ✅ Horizontal scaling support

### 13.5 Phase 5: Ecosystem (In Progress)

**Planned:**
- 🔄 Marketplace for templates
- 🔄 Custom integration SDK
- 🔄 CLI tool for workflow management
- 🔄 VS Code extension
- ⏳ Terraform provider
- ⏳ GitHub Action for CI/CD

### 13.6 Phase 6: Enterprise Plus (Planned)

**Planned:**
- ⏳ SSO (SAML, OIDC)
- ⏳ Advanced RBAC (attribute-based)
- ⏳ Data residency controls
- ⏳ SOC 2 compliance
- ⏳ SLA tiers with priority execution
- ⏳ White-label deployment

### 13.7 Phase 7: AI-Native (Future)

**Planned:**
- ⏳ Natural language workflow creation
- ⏳ AI-powered debugging
- ⏳ Anomaly detection
- ⏳ Smart recommendations
- ⏳ Auto-remediation

---

## 14. Risks & Mitigations

### 14.1 Technical Risks

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| **Database bottleneck** | High | Medium | Read replicas, caching, sharding strategy |
| **Integration failures** | High | High | Circuit breakers, fallbacks, retry logic |
| **Security vulnerability** | Critical | Medium | Regular audits, bug bounty, dependency scanning |
| **Performance degradation** | Medium | Medium | Load testing, auto-scaling, performance budgets |
| **Data loss** | Critical | Low | WAL archiving, point-in-time recovery, geo-replication |

### 14.2 Product Risks

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| **Complex onboarding** | High | Medium | Templates, guided tours, documentation |
| **Feature bloat** | Medium | Medium | User research, feature flags, analytics |
| **Integration gaps** | Medium | High | Community contributions, integration SDK |
| **Competitor catch-up** | Medium | High | Innovation velocity, community building |

### 14.3 Business Risks

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| **Slow enterprise adoption** | High | Medium | Case studies, POC support, enterprise features |
| **Community fragmentation** | Medium | Low | Governance, contributor guidelines, roadmap transparency |
| **Funding constraints** | High | Medium | Open-core model, early enterprise sales |
| **Key person dependency** | Medium | Medium | Documentation, knowledge sharing, team growth |

### 14.4 Compliance Risks

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| **GDPR violation** | Critical | Low | Data residency, right to deletion, DPO |
| **SOC 2 gap** | Medium | Medium | Control implementation, audit prep |
| **Data breach** | Critical | Low | Encryption, access controls, incident response |

---

## 15. Glossary

| Term | Definition |
|------|------------|
| **Action** | A single step in a workflow that performs an operation |
| **CEL** | Common Expression Language used for dynamic values |
| **Circuit Breaker** | Pattern to prevent cascading failures from integration errors |
| **DAG** | Directed Acyclic Graph - workflow structure without cycles |
| **DEK** | Data Encryption Key - encrypts credential data |
| **DLQ** | Dead Letter Queue - stores failed messages for retry |
| **Edge** | Connection between two nodes in a workflow |
| **Execution** | A single run of a workflow |
| **Human Task** | A workflow step requiring manual approval |
| **KEK** | Key Encryption Key - encrypts DEKs |
| **Node** | A single element in a workflow (trigger, action, condition) |
| **RBAC** | Role-Based Access Control |
| **Schedule** | Cron-based trigger for automated workflow execution |
| **Step** | Individual node execution within a workflow run |
| **Sub-Workflow** | A workflow invoked by another workflow |
| **Tenant** | Isolated organizational unit with separate data |
| **Trigger** | Entry point that starts a workflow execution |
| **Visitor Pattern** | Design pattern used for workflow execution traversal |
| **Webhook** | HTTP endpoint that triggers workflow execution |
| **Workflow** | A defined sequence of automated actions |

---

## 16. Appendices

### Appendix A: Integration Catalog

| Category | Integration | Actions |
|----------|-------------|---------|
| **Communication** | Slack | send_message, send_dm, add_reaction, update_message |
| | Email (SendGrid) | send_email, send_template |
| | Email (SES) | send_email |
| | SMS (Twilio) | send_sms |
| **Issue Tracking** | GitHub | create_issue, add_label, create_comment |
| | Jira | create_issue, transition, add_comment, search |
| | PagerDuty | create_incident, acknowledge, resolve |
| **Cloud** | AWS Lambda | invoke |
| | AWS SNS | publish |
| | AWS SQS | send_message |
| | AWS S3 | get_object, put_object |
| **AI/LLM** | OpenAI | chat_completion, embedding |
| | Anthropic | chat_completion |
| | Bedrock | invoke_model |
| **Utility** | HTTP | request (GET, POST, PUT, DELETE) |
| | JavaScript | execute |
| | Transform | map, filter, reduce |

### Appendix B: Expression Examples

```
# Access trigger payload
{{trigger.payload.user.email}}

# Access step output
{{steps.http_request.response.body.id}}

# Conditional expression
{{trigger.payload.priority == "high" ? "urgent" : "normal"}}

# Array operations
{{trigger.payload.items.map(i => i.name)}}

# Environment variables
{{env.SLACK_CHANNEL}}

# Credential reference
{{credentials.github_token}}

# Built-in functions
{{now()}}
{{uuid()}}
{{base64Encode(trigger.payload.data)}}
```

### Appendix C: Webhook Signature Verification

```go
// Verify webhook signature
func VerifyWebhookSignature(payload []byte, signature, secret string) bool {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(payload)
    expected := hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(signature), []byte(expected))
}
```

**Header Format:**
```
X-Gorax-Signature: sha256=abc123...
X-Gorax-Timestamp: 1234567890
X-Gorax-Delivery-ID: del_xyz789
```

### Appendix D: Rate Limits

| Endpoint | Limit | Window |
|----------|-------|--------|
| `/api/v1/*` | 1000 | 1 minute |
| `/webhook/*` | 5000 | 1 minute |
| `/api/v1/workflows/*/execute` | 100 | 1 minute |
| `/api/v1/credentials` | 50 | 1 minute |

**Response Headers:**
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 950
X-RateLimit-Reset: 1234567890
```

### Appendix E: Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | Required |
| `REDIS_URL` | Redis connection string | Required |
| `SQS_QUEUE_URL` | AWS SQS queue URL | Required |
| `KMS_KEY_ID` | AWS KMS key for encryption | Required (prod) |
| `KRATOS_URL` | Ory Kratos URL | Required (prod) |
| `SENTRY_DSN` | Sentry error tracking | Optional |
| `OTEL_ENDPOINT` | OpenTelemetry collector | Optional |
| `LOG_LEVEL` | Logging verbosity | `info` |
| `PORT` | HTTP server port | `8080` |

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-12 | Engineering Team | Initial comprehensive PRD |

---

## Next Steps

1. **Review and Approval**: Share with stakeholders for feedback
2. **Technical Design**: Create detailed technical specifications for Phase 5 features
3. **Sprint Planning**: Break down Phase 5 into sprint-sized deliverables
4. **Community Input**: Publish roadmap for community feedback

---

*This is a living document and will be updated as the product evolves.*
