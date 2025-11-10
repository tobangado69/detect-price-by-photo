# Fullstack PRD: Detect Price by Photo

## 1. Executive Summary

**Product Name:** Detect Price by Photo

**Version:** 1.0 MVP

**Status:** In Development

**Target Launch:** Q1 2026

**Target Market:** Indonesia

**Primary Goal:** Provide Indonesian online sellers with an AI-powered, mobile-first SaaS platform to instantly estimate accurate market prices for products using image analysis and real-time market data.

**Key Value Proposition:**
- **Speed:** Get price estimates in <2 seconds (vs. 10+ minutes manual research)
- **Accuracy:** 85%+ confidence scores backed by market data and RAG
- **Cost-Effective:** Free tier available, Premium at Rp 65K/month with 1,000 daily analyses
- **Localized:** Indonesian language, IDR currency, local payment methods (Midtrans)

---

## 2. Business Model & Revenue Strategy

### 2.1 Subscription Tiers

| Tier | Daily Limit | Monthly Price (IDR) | Annual Price (IDR) | Target Users |
|------|-----------|-------------------|------------------|------------|
| **Free** | 10 | 0 | 0 | Casual traders, students (CAC: Organic) |
| **Premium** | 1,000 | 65,000 | 650,000 (-20% discount) | Active sellers, resellers (CAC: ~Rp 50K/user) |
| **Enterprise** | Unlimited | Custom | Custom | Large merchants, B2B integrations (CAC: Direct sales) |

### 2.2 Financial Projections (12-Month)

| Metric | Month 1 | Month 6 | Month 12 |
|--------|--------|--------|---------|
| Total Users | 500 | 15,000 | 50,000 |
| Premium Users | 20 | 2,000 | 12,000 |
| MRR (Rp millions) | 1.3 | 130 | 780 |
| CAC (Rp) | 0 | 25,000 | 30,000 |
| LTV (Rp) | 0 | 1,170,000 | 1,400,000 |
| COGS (as % of revenue) | 60% | 40% | 30% |

**COGS Breakdown:**
- OpenRouter API: ~30-40% of revenue (primary cost)
- AWS S3/CDN: 3-5%
- Midtrans fees: 2.5-3%
- Infrastructure: 5-8%
- Support/Operations: 10-15%

### 2.3 Customer Acquisition Strategy

**Phase 1 (Months 1-3):**
- Direct outreach to marketplace seller communities
- Content marketing (YouTube, TikTok tutorials)
- Influencer partnerships with e-commerce creators
- Referral program (Rp 10K per referral)

**Phase 2 (Months 4-9):**
- Paid ads (Google Ads, Meta Ads) targeting seller keywords
- Marketplace partnerships (sponsorships on Tokopedia, Shopee)
- Integration partnerships (listing tools, inventory management)

**Phase 3 (Months 10+):**
- B2B sales to large merchants
- White-label/API tier for enterprise customers

---

## 3. Complete User Journey Map

### 3.1 New User Journey (Day 1-7)

```
Sign Up → Email Verification → Onboarding Tour → Upload Photo 
  ↓
Free Tier (10 daily) → First Estimate → History View → Share Result
  ↓
Quota Warning (9/10 used) → Upgrade Prompt → Subscribe to Premium
```

**Key Touchpoints:**
- **Day 1:** Welcome email with tutorial video link
- **Day 3:** Reminder email if haven't uploaded yet ("Try our AI appraiser")
- **Day 7:** Engagement metric (if used 7+ times → high retention signal)

### 3.2 Premium User Journey (Retention)

```
Premium Activated → Bulk Upload Feature → API Access → Analytics Dashboard
  ↓
Monthly Usage Report → Insights ("You saved 50 hours this month") → Renewal
```

### 3.3 Churn Prevention

| Churn Signal | Action | Trigger |
|---|---|---|
| No login for 7 days | Re-engagement email + "Miss us?" | Manual |
| Usage dropped 50% vs last month | Check-in email + feature tips | Automated |
| Support ticket filed | Prioritize response, offer discount | Manual |
| Premium user cancels | Exit survey + 20% retention offer | Manual |

---

## 4. System Architecture Overview

### 4.1 High-Level System Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                    Frontend Layer                           │
│  Vite + React + TypeScript + Shadcn UI + Tailwind CSS       │
│  (Mobile/Tablet/Desktop - Responsive)                       │
└──────────────────────────┬──────────────────────────────────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
        ▼                  ▼                  ▼
┌─────────────────┐  ┌──────────────┐  ┌──────────────┐
│   Auth API      │  │ Analysis API │  │ Admin API    │
│   (JWT/Refresh) │  │ (Estimate)   │  │ (Management) │
└─────────────────┘  └──────────────┘  └──────────────┘
        │                  │                  │
        └──────────────────┼──────────────────┘
                           ▼
        ┌──────────────────────────────────┐
        │  API Gateway / Load Balancer     │
        │  - Rate limiting (per-plan)      │
        │  - Request validation            │
        │  - Logging & monitoring          │
        └──────────────────────┬───────────┘
                               ▼
        ┌──────────────────────────────────────────┐
        │  Backend (Golang) - Raw HTTP + SQL       │
        │  - HTTP server (net/http)                │
        │  - Business logic layer                  │
        │  - Service layer (Auth, Estimation, etc) │
        │  - RAG pipeline                          │
        └──────────┬──────────────┬────────────────┘
                   │              │
    ┌──────────────┴────┐  ┌─────┴──────────┐
    │                   │  │                │
    ▼                   ▼  ▼                ▼
┌──────────┐  ┌─────────────────┐  ┌────────────┐
│PostgreSQL│  │  pgvector +     │  │   Redis    │
│(Users,   │  │ Market Knowledge│  │ (Cache)    │
│Billing,  │  │ Base (Embeddings)│  │ (Sessions) │
│Analysis) │  └─────────────────┘  └────────────┘
└──────────┘
    │
    ├──────────────────────────────────────────┐
    │                                          │
    ▼                                          ▼
┌─────────────────┐              ┌──────────────────┐
│  OpenRouter API │              │ Midtrans Payment │
│  (400+ models)  │              │ (Credit/GoPay)   │
└─────────────────┘              └──────────────────┘
```

### 4.2 Data Flow: Photo Upload → Price Estimate

```
1. User Upload (Frontend)
   └─> PhotoUploadCard component
       └─> Validate file (size, format)
       └─> Compress image
       └─> Upload to S3
       
2. Backend Receives Request
   └─> Authenticate user (JWT)
   └─> Check daily quota
   └─> Queue analysis job
   
3. AI Estimation Logic
   ├─> Decision Tree:
   │   ├─> FAST mode: Hash-based cache lookup (2-5s)
   │   ├─> ACCURATE mode: Full LLM call (15-25s)
   │   └─> KNOWLEDGE mode: RAG + LLM (10-20s)
   │
   └─> RAG Pipeline:
       ├─> Extract embedding: product_name + condition
       ├─> Vector search in pgvector (cosine similarity)
       ├─> Retrieve top-5 market records
       ├─> Pass to LLM with context
       └─> LLM generates price + reasoning
       
4. Store & Cache
   └─> Save to PostgreSQL (price_analyses table)
   └─> Cache in Redis for 24 hours
   └─> Decrement daily quota
   
5. Return Results
   └─> API response with price range + confidence
   └─> Frontend displays results
   └─> Optional: Share via WhatsApp/SMS
```

### 4.3 Database Schema (Core Tables)

```
users (id, email, password_hash, first_name, company, status, created_at)
  └─ One-to-Many: subscriptions
  └─ One-to-Many: price_analyses
  └─ One-to-Many: payments

subscriptions (id, user_id, plan_id, status, daily_limit, usage_today, reset_at)
  └─ Many-to-One: users
  └─ Many-to-One: plans

price_analyses (id, user_id, image_url, product_name, condition, 
                estimated_price, confidence_score, created_at)
  └─ Many-to-One: users

market_knowledge (id, product_name, condition, embedding, market_data)
  └─ Vector column with HNSW index

plans (id, name, daily_limit, price_idr, features)

payments (id, user_id, transaction_id, amount, status, created_at)
  └─ Many-to-One: users
```

---

## 5. Feature Priority Matrix

### 5.1 MVP Launch (Weeks 1-8)

**Must-Have (P0 - Critical):**
- [ ] User registration & login (JWT + refresh tokens)
- [ ] Single photo upload with preview
- [ ] Price estimation API integration (OpenRouter)
- [ ] Result display with confidence score
- [ ] Free tier (10 daily analyses)
- [ ] Premium tier upgrade flow (Midtrans integration)
- [ ] Invoice generation & download
- [ ] History view (past analyses)
- [ ] Mobile-responsive design
- [ ] Error handling & user feedback
- [ ] Basic security (input validation, HTTPS)

**Should-Have (P1 - High Priority):**
- [ ] Admin dashboard (user management, analytics)
- [ ] RAG market knowledge base setup
- [ ] Multiple AI model support (fast/accurate/knowledge modes)
- [ ] Email notifications (welcome, invoice, renewal)
- [ ] Accessibility compliance (WCAG 2.1 AA)
- [ ] Performance optimization (sub-2s response time)

### 5.2 Phase 2 (Weeks 9-16)

- Batch upload feature
- Advanced analytics for users
- B2B API tier with documentation
- Integration with marketplace APIs (Tokopedia, Shopee)
- Referral program

### 5.3 Phase 3 (Months 6+)

- Mobile app (iOS/Android)
- Predictive pricing recommendations
- AI-powered bulk pricing suggestions
- White-label/SaaS reseller program

---

## 6. Technology Stack Decision Matrix

| Component | Choice | Rationale | Alternatives Considered |
|-----------|--------|-----------|------------------------|
| **Backend** | Golang + stdlib | Fast, minimal deps, great concurrency, startup-friendly | Node.js, Python |
| **Frontend** | Vite + React | Modern, fast builds, best ecosystem for SaaS UIs | Vue, Svelte, Next.js |
| **Database** | PostgreSQL | Relational data + pgvector for RAG, single system | MongoDB, MySQL |
| **Cache** | Redis | In-memory, fast, supports TTL, rate limiting | Memcached, DynamoDB |
| **AI/LLM** | OpenRouter | 400+ models, cost optimization, flexible switching | Direct OpenAI, Anthropic |
| **Payments** | Midtrans | Indonesia-focused, QRIS/GoPay support, competitive fees | Stripe, PayPal |
| **Storage** | AWS S3 | Reliable, CDN integration, Indonesia region available | GCS, DigitalOcean |
| **Container** | Docker | Dev/prod parity, easy deployment | VM images, serverless |

---

## 7. Security & Compliance

### 7.1 Data Protection

- **Encryption in transit:** TLS 1.3+ for all API traffic
- **Encryption at rest:** PostgreSQL encryption, S3 server-side encryption
- **Password hashing:** bcrypt with salt (cost factor 12)
- **Token management:** JWT access tokens (15min) + refresh tokens (7 days)
- **PCI-DSS:** Payments via Midtrans (no card data stored locally)

### 7.2 Regulatory Compliance

- **Indonesia GDPR (PDP Law):** Data residency, user consent management
- **GDPR (if EU users):** Data export, right to deletion, consent tracking
- **SOC 2 (future):** Security audit and compliance certification

### 7.3 API Security

- **CORS:** Whitelist production domains only
- **Rate limiting:** 100 req/min per user, escalating for auth endpoints
- **SQL injection:** Prepared statements for all queries
- **XSS prevention:** Input sanitization, React's built-in escaping
- **CSRF:** Token-based protection for state-changing operations

---

## 8. Operational Plan

### 8.1 Development Timeline (16 Weeks)

**Phase 1: Weeks 1-4 (Foundation)**
- Database schema & migrations
- Backend API scaffolding (auth, users, subscriptions)
- Frontend login/register pages
- Midtrans payment integration setup

**Phase 2: Weeks 5-8 (Core Feature)**
- Photo upload & S3 integration
- OpenRouter API integration
- Price estimation logic & RAG setup
- Result display & history view
- Admin dashboard (basic)

**Phase 3: Weeks 9-12 (Polish & Testing)**
- Performance optimization
- Bug fixes & QA testing
- Accessibility compliance
- Security audit & penetration testing
- Documentation & API specs

**Phase 4: Weeks 13-16 (Deployment & Beta)**
- CI/CD pipeline setup
- Production infrastructure deployment
- Beta testing with 100-500 selected users
- Final polish based on beta feedback
- Marketing materials preparation

### 8.2 Team Structure (Recommended)

| Role | Count | Responsibilities |
|------|-------|-----------------|
| Backend Developer | 1 | Golang API, database, OpenRouter integration |
| Frontend Developer | 1 | React UI, mobile responsiveness, state management |
| Product Manager | 1 | Requirements, prioritization, user research |
| DevOps/Infra | 0.5 | Docker, CI/CD, AWS, monitoring |
| QA Engineer | 0.5 | Testing, bug reporting, test automation |
| **Total** | **3-4** | **Full team to launch MVP** |

### 8.3 Resource Budget (12-Month)

| Item | Monthly | Annual |
|------|---------|--------|
| **Infrastructure** | | |
| AWS (EC2, RDS, S3) | $2,000 | $24,000 |
| OpenRouter AI credits | $3,000 | $36,000 |
| Redis/CDN | $500 | $6,000 |
| **People** | | |
| 3-4 developers (avg) | $15,000 | $180,000 |
| **Tools & Services** | | |
| GitHub, monitoring, etc. | $500 | $6,000 |
| **Marketing/Sales** | $2,000 | $24,000 |
| **Total** | **$23,000** | **$276,000** |

---

## 9. Marketing & Growth Strategy

### 9.1 Go-to-Market (GTM)

**Phase 1: Organic (Months 1-3)**
- Free tier as acquisition channel (viral loop: referrals)
- YouTube tutorial videos (how to price products)
- TikTok content with marketplace sellers (UGC)
- Influencer partnerships (micro-influencers: 10K-100K followers)
- Community engagement (Reddit, Facebook groups for Indonesian sellers)

**Phase 2: Paid (Months 4-9)**
- Google Ads (keywords: "harga barang bekas", "penilaian harga jual")
- Meta Ads (retargeting website visitors)
- Marketplace sponsorships (featured app on Tokopedia, Shopee)

**Phase 3: Strategic Partnerships (Months 10+)**
- Integrations with seller tools (inventory management, listing tools)
- Partnerships with logistics/fulfillment providers
- B2B sales to large merchants

### 9.2 Key Metrics & KPIs

| Metric | Month 3 Target | Month 12 Target |
|--------|---|---|
| **Acquisition** | | |
| Total signups | 3,000 | 50,000 |
| Daily active users (DAU) | 500 | 8,000 |
| Premium conversion rate | 2% | 25% |
| **Engagement** | | |
| Avg analyses per user/month | 15 | 25 |
| User retention (30-day) | 40% | 70% |
| **Revenue** | | |
| MRR | Rp 2M | Rp 780M |
| Premium subscribers | 50 | 12,000 |
| **Efficiency** | | |
| CAC | Rp 0 (organic) | Rp 30K |
| LTV | Rp 500K | Rp 1.4M |
| CAC Payback Period | N/A | 1.3 months |

---

## 10. Risk Assessment & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|-----------|
| **AI Model Inaccuracy** | High | Medium | RAG refinement, user feedback loops, manual review for disputes, A/B testing models |
| **Midtrans Downtime** | High | Low | Fallback payment methods (manual bank transfer), queue system for retries |
| **OpenRouter API Costs Spike** | Medium | Medium | Cost monitoring, model switching to cheaper alternatives, usage-based quotas |
| **Market Competition** | Medium | Medium | Focus on Indonesia market specificity, local partnerships, continuous improvement |
| **Regulatory Changes** | Medium | Low | Legal compliance review quarterly, stay updated on PDP Law changes |
| **Churn > 10%** | High | Medium | Engagement campaigns, feature releases, loyalty discounts |
| **Scaling Infrastructure Issues** | High | Low | Load testing (Month 3), auto-scaling setup, database optimization |

---

## 11. Success Criteria & Launch Checklist

### 11.1 Beta Launch Criteria

- [ ] All P0 features implemented & tested
- [ ] API response time P95 <2 seconds
- [ ] Price estimation accuracy >85% (test with 100 items)
- [ ] 99.5% uptime during 2-week beta
- [ ] Midtrans integration fully tested (sandbox & production)
- [ ] Mobile responsive (tested on iOS 15+, Android 11+)
- [ ] Security audit passed (no critical vulnerabilities)
- [ ] Documentation complete (API docs, user guide, admin guide)
- [ ] 100-500 beta users recruited & informed
- [ ] Feedback collection system set up (Typeform, in-app surveys)

### 11.2 Public Launch Criteria

- [ ] Beta feedback incorporated (priority fixes completed)
- [ ] Marketing materials ready (landing page, video, social posts)
- [ ] Press release prepared & sent to tech media
- [ ] Analytics tracking set up (Plausible/Mixpanel)
- [ ] Error tracking configured (Sentry)
- [ ] Support team trained & on-call
- [ ] Load testing passed (10K concurrent users)
- [ ] Disaster recovery plan documented & tested

---

## 12. Post-Launch Support & Operations

### 12.1 SLA & Support

- **Response time:** Support tickets within 24 hours (working days)
- **Bug fix:** Critical bugs within 4 hours, High within 24 hours
- **Uptime SLA:** 99.5% (12.2 hours downtime/month permitted)
- **Support channels:** Email, in-app chat, WhatsApp (for Premium/Enterprise)

### 12.2 Continuous Improvement

**Monthly Reviews:**
- User feedback analysis
- Feature request prioritization
- Performance metrics review
- Churn analysis & retention initiatives

**Quarterly Goals:**
- New feature releases
- Infrastructure optimization
- Security updates
- Market expansion planning

---

## 13. Appendix: Integration Points

### 13.1 OpenRouter API Integration

```
Model Selection Logic:
- Fast mode: GPT-4o-mini (cheaper, 2-5s)
- Accurate mode: GPT-4o (best quality, 15-25s)
- Knowledge mode: Claude-3.5-sonnet (excellent reasoning, 10-20s)

Fallback: If primary API fails → Retry with fallback model
Cost Optimization: Track tokens/cost per request, alert if >Rp 100 per estimate
```

### 13.2 Midtrans Payment Flow

```
1. User clicks "Upgrade" → Frontend creates payment request
2. Backend calls Midtrans API → Get snap token
3. Frontend shows Midtrans payment page (iframe)
4. User completes payment (credit card, GoPay, QRIS)
5. Midtrans sends webhook → Backend updates subscription
6. User redirected back to app with success message
```

### 13.3 PostgreSQL + pgvector Setup

```
Market Knowledge Table:
- product_name: "iPhone 13 Pro Max"
- condition: "good"
- embedding: [0.123, -0.456, ...] (1536-dim vector)
- market_data: JSON with pricing stats
- Index: HNSW on embedding column for fast similarity search
```

---

## Conclusion & Next Steps

The **Detect Price by Photo** platform is positioned to capture significant market share in Indonesia's growing e-commerce sector. By focusing on a specific user need (accurate, fast price estimation), leveraging advanced AI/RAG technology, and maintaining a lean operational structure, the platform can achieve profitability within 12-18 months.

**Immediate Next Steps:**
1. Form dev team & assign owners
2. Set up development environment (GitHub, Docker, CI/CD)
3. Begin Phase 1 (database & API scaffolding)
4. Recruit beta testers from seller communities
5. Start content creation for marketing

---

**Document Version:** 1.0  
**Last Updated:** 2025-11-08  
**Next Review Date:** 2025-12-08