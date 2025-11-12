---
title: Frontend PRD
weight: 2
---

# Frontend PRD: Detect Price by Photo

## 1. Executive Summary

**Product Name:** Detect Price by Photo (Frontend Application)

**Version:** 1.0 MVP

**Status:** In Development

**Target Release:** Q1 2026

**Platform:** Web (Desktop & Mobile Responsive)

**Technology Stack:** Vite.js + React + TypeScript + Shadcn UI + Tailwind CSS

This document defines the frontend user interface, user experience flows, component architecture, and feature set for the **Detect Price by Photo** application. The frontend is the primary touchpoint for Indonesian online sellers to upload product photos and receive AI-powered price estimates.

---

## 2. Product Vision & UX Principles

### 2.1 Vision Statement

Create an intuitive, mobile-first interface that enables users to upload product photos and receive accurate price estimates within seconds, making data-driven pricing decisions effortless.

### 2.2 UX Principles

| Principle | Description |
|-----------|-------------|
| **Simplicity** | 3 clicks to upload photo and get estimate (mobile), no friction |
| **Transparency** | Clear confidence scores, reasoning, market data visualization |
| **Performance** | Sub-2 second response times, smooth animations, no loading frustration |
| **Localization** | Bahasa Indonesia interface, IDR currency, local payment methods |
| **Accessibility** | WCAG 2.1 AA compliance, keyboard navigation, screen reader support |
| **Trust** | Show real market data, user testimonials, accuracy metrics |

### 2.3 Design System

**Color Palette (Tailwind):**
- Primary: `#3B82F6` (Blue - Action, trust)
- Success: `#10B981` (Green - Positive price estimates)
- Warning: `#F59E0B` (Amber - Low confidence, quota warnings)
- Danger: `#EF4444` (Red - Errors, failed uploads)
- Neutral: `#6B7280` (Gray - Text, secondary)

**Typography:**
- Headings: Inter Bold/SemiBold
- Body: Inter Regular
- Mono: IBM Plex Mono (code snippets)

---

## 3. Target Users & User Personas

### 3.1 Primary User Persona

**Agam, 28 - Online Reseller**
- Sells 50-100 items daily on marketplace platforms
- Frequent user (3-4x per day)
- Limited technical knowledge, prefers simple UI
- Uses mobile 80% of the time
- Value: Speed and accuracy
- Pain point: Spending 10+ minutes per item pricing

**Siti, 35 - Marketplace Manager**
- Manages teams selling across multiple platforms
- Batch uploader (wants bulk features)
- Occasional user (3-5x per week)
- Comfortable with APIs and integrations
- Value: Efficiency, API access, reporting
- Pain point: Manual pricing for 100+ items weekly

**Budi, 22 - Casual Trader**
- Sells used items occasionally (2-3x per month)
- Free tier user (price-sensitive)
- Mobile-first, social-media native
- First time in e-commerce
- Value: Free tier availability, simple process
- Pain point: Doesn't know current market prices

### 3.2 Secondary Users

**Admin/Support Team:**
- Internal dashboard access
- Manages user disputes, subscription issues
- Limited interaction with main app

---

## 4. Key Features & User Flows

### 4.1 Authentication & Onboarding

#### User Registration
1. User clicks "Sign Up" → Email/Phone form
2. Enter password (strength indicator in real-time)
3. Verify email (OTP sent)
4. Set profile (name, company optional)
5. Auto-enrolled in Free tier
6. Quick onboarding tour (2-3 slides)

**Components:**
- `AuthPage` (registration/login form)
- `PasswordStrengthIndicator` (real-time feedback)
- `EmailVerificationModal` (OTP input)
- `OnboardingTour` (carousel with skip option)

#### Login
- Email + Password
- Remember me (30 days)
- Forgot password flow
- Social login (future: Google, Facebook)

### 4.2 Main Feature: Photo Upload & Price Estimation

#### Flow A: Single Upload (90% of usage)

```
1. User clicks "Upload Photo" → Camera capture modal opens (uses `navigator.mediaDevices.getUserMedia`)
2. Live preview displayed with capture button; user can:
   - Take a new photo directly in browser (preferred)
   - Or choose an existing file via fallback file input (desktop compatibility)
3. Captured frame is converted to `Blob`/`File` and auto-previewed
4. User enters:
   - Product name (searchable autocomplete, e.g., "iPhone 13")
   - Condition selector (radio: new/like_new/good/fair/poor)
   - Optional: Detailed description
   - Mode selector (advanced users): fast/accurate/knowledge
5. Click "Estimate Price"
6. Loader animation (estimated 2-5 seconds)
7. Results displayed with:
   - Estimated price range (min-median-max)
   - Confidence score (0-100%)
   - Market insights (avg listings, price trend)
   - Reasoning explanation
   - "Share" button (SMS, WhatsApp, social)
```

**Components:**
- `CameraCaptureModal` (controls permission request, live preview, capture)
- `PhotoUploadCard` (drag-drop area + file input fallback)
- `PhotoPreview` (image display + crop tool)
- `ProductNameInput` (autocomplete from market data)
- `ConditionSelector` (visual radio buttons + descriptions)
- `PriceEstimateResult` (card with visualization)
- `PriceChart` (simple range bar + market comparison)
- `LoadingAnimation` (cute Lottie animation)
- `ConfidenceScore` (visual gauge + explanation)

#### Flow B: Batch Upload (Premium feature)

```
1. User clicks "Batch Upload"
2. Upload 1-20 photos at once
3. System prompts for common product name (or bulk CSV)
4. Queue displayed (upload progress)
5. Results list view with export options
```

### 4.3 History & Analytics

**User Dashboard - Main Tab:**
- Card showing: "Analyses This Month: 342 / 1000"
- Progress bar with upgrade CTA if nearing limit
- Search/filter by date, product, confidence
- Recent analyses in scrollable list/grid
- Each item shows: thumbnail, product, estimated price, date, quick re-analyze button

**Components:**
- `HistoryList` (table with sorting, pagination)
- `AnalysisCard` (compact result view)
- `QuotaProgressBar` (visual indicator + upgrade button)
- `ExportButton` (CSV download)

### 4.4 Subscription & Billing

**Pricing Page:**
- 3-column layout: Free | Premium | Enterprise
- Feature comparison table below
- Pricing in IDR with annual discount highlight (e.g., "Save 20%")
- "Upgrade" CTA button on Premium/Enterprise
- FAQ accordion

**Subscription Management Page:**
- Current plan card with features list
- "Upgrade" / "Cancel" buttons
- Invoice history (downloadable PDFs)
- Payment method on file
- Billing cycle and next renewal date

**Payment Flow:**
1. User clicks "Upgrade to Premium"
2. Modal shows plan details + Midtrans payment methods
3. Select: Credit/Debit card, GoPay, QRIS
4. Redirect to Midtrans payment gateway (iframe or redirect)
5. On success: Return to app, show "Subscription Activated!" toast
6. New limits reflected immediately

**Components:**
- `PricingCard` (feature list + CTA)
- `SubscriptionCard` (current plan details)
- `PaymentMethodSelector` (radio buttons with icons)
- `InvoiceTable` (history + download links)
- `UpgradeModal` (plan confirmation + payment)

### 4.5 Navigation & Layout

**Main Layout:**
```
┌─────────────────────────────────┐
│  Logo    Search    User Menu    │  <- Header (sticky)
├─────────────────────────────────┤
│   │                             │
│ S │    MAIN CONTENT             │
│ i │    (dynamic based on page)   │
│ d │                             │
│ e │                             │
│ b │                             │
│ a │                             │
│ r │                             │
│   │                             │
├─────────────────────────────────┤
│  Footer (links, social)         │  <- Footer
└─────────────────────────────────┘
```

**Sidebar (Desktop - Toggle on Mobile):**
- Home (Dashboard)
- Upload (Primary CTA)
- History
- Subscription
- Settings
- Help & FAQ
- Admin (if admin user)

**Mobile Navigation:**
- Bottom tab bar (5 main sections)
- Hamburger menu for secondary items

**Components:**
- `AppLayout` (wrapper with sidebar + main content)
- `Sidebar` (collapsible, responsive)
- `Header` (logo, search, user dropdown, `Iconify` icons for actions)
- `BottomNav` (mobile-only, `Iconify` icons for navigation)
- `Footer` (links, social icons powered by `Iconify`)

---

## 5. Page Structure & Component Breakdown

### 5.1 Page: Home / Dashboard

**Purpose:** Show quick stats, CTA to upload, recent history

**Content:**
- Hero section: "Quick price estimation" + upload button
- Stats cards: Total analyses, avg confidence, photos remaining
- Recent analyses (6-item grid)
- "Learn more" link to tutorial video
- Upgrade prompt if Free tier user

**Components Used:**
- `HeroSection`
- `StatsCard` (4-card grid)
- `AnalysisGrid`
- `UpgradePromoBanner`

### 5.2 Page: Upload / Analyze

**Purpose:** Primary feature - upload and analyze

**Content:**
- Large upload area (drag-drop + button)
- Form: Product name, condition, mode
- Preview of uploaded image
- Loading state with percentage + estimated time
- Results view (once analysis complete)

**Components Used:**
- `PhotoUploadCard` (uses `Iconify` for button glyphs)
- `ProductForm`
- `LoadingAnimation`
- `ResultsView`

### 5.3 Page: History

**Purpose:** View all past analyses, search, filter, export

**Content:**
- Filters: Date range, product, confidence score range
- Table view: Image thumbnail, product, price, confidence, date, actions
- Pagination / infinite scroll
- Export as CSV button

**Components Used:**
- `FilterBar`
- `HistoryTable`
- `ImageThumbnail`
- `Pagination`

### 5.4 Page: Subscription

**Purpose:** Manage subscription, upgrade, view invoices

**Content:**
- Current plan card (name, features, renewal date)
- Pricing comparison table
- Invoice history (searchable, downloadable)
- Payment method management

**Components Used:**
- `SubscriptionCard`
- `PricingTable`
- `InvoiceTable`
- `PaymentMethodForm`

### 5.5 Page: Settings

**Purpose:** User preferences, profile, security

**Content:**
- Profile: Name, email, phone, company
- Password change
- Email preferences (notifications)
- Language selection (future: multi-language)
- Delete account option

**Components Used:**
- `ProfileForm`
- `PasswordChangeForm`
- `NotificationPreferences`
- `DangerZone` (delete account)

### 5.6 Page: Admin Dashboard (Admin Only)

**Purpose:** Manage users, subscriptions, model configuration

**Content (Tabs):**
1. **Users Tab:** List all users, search, filter by plan, suspend/upgrade
2. **Subscriptions Tab:** Revenue metrics, plan distribution, retention
3. **Payments Tab:** Transaction log, failed payments, disputes
4. **AI Models Tab:** Select global default model, configure routing, view costs
5. **Analytics Tab:** Charts - DAU, revenue trend, accuracy metrics

> Only administrators see the AI Models tab. End users always run analyses against the admin-defined default model; the UI never exposes model selection to them.

**Components Used:**
- `UserTable` (admin features)
- `RevenueChart`
- `ModelConfigPanel` (dropdown of allowed models, set-default CTA, routing editor)
- `AnalyticsBoard`

---

## 6. State Management & Data Flow

### 6.1 State Management Strategy

Use **Zustand** for global state (lightweight, easier than Redux):

**Stores:**
1. **AuthStore** - user, token, isAuthenticated
2. **SubscriptionStore** - currentPlan, dailyLimit, usageToday
3. **AnalysesStore** - recentAnalyses, filter state
4. **UIStore** - sidebarOpen, darkMode, locale

### 6.2 API Calls via Axios

**API Client Setup:**
```typescript
// services/api.ts
const apiClient = axios.create({
  baseURL: 'https://api.detectpricebyphoto.com/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Auto-attach JWT token to requests
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Handle token refresh on 401
apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      const newToken = await refreshAuthToken();
      // Retry request with new token
    }
    return Promise.reject(error);
  }
);
```

### 6.3 Data Flow Example: Upload & Estimate

```
User Action → UI Component → Zustand Store → API Call → Backend
  ↓              ↓                ↓               ↓
Upload       PhotoUpload      Set loading       POST
Photo        Component        state             /analyses
  ↓                           ↓                 /estimate
             onPhotoSelect()   dispatch()
                              action            Response
                                                ↓
                              ← Parse response  ← Return
                                                  estimate
                              Store analysis     ↓
                              Set results UI
```

---

## 7. Component Library & Shadcn UI Integration

### 7.1 Shadcn UI Components Used

| Component | Usage |
|-----------|-------|
| **Button** | CTAs, form submission, navigation |
| **Input** | Text fields (email, password, product name) |
| **Select** | Dropdowns (condition, mode, plan selection) |
| **Card** | Content containers (stats, results, pricing) |
| **Dialog/Modal** | Confirmation, payment modal, forms |
| **Table** | Admin tables, history view, invoices |
| **Progress** | Quota bar, upload progress |
| **Tabs** | Admin dashboard sections |
| **Badge** | Status labels (active, pending, etc.) |
| **Toast** | Notifications (success, error, info) |
| **Avatar** | User profile picture, dropdown menu |
| **Slider** | Price range filter |
| **Switch** | Toggles (dark mode, notifications) |

### 7.2 Custom Components (Not in Shadcn)

- `PhotoUploadCard` - Drag-drop file upload with preview
- `ConfidenceGauge` - Visual confidence score (0-100%)
- `PriceRangeVisualization` - Bar chart with min/med/max
- `LoadingAnimation` - Lottie animation during processing
- `MarketInsightsPanel` - Market data visualization
- `QuotaProgressBar` - Daily usage progress indicator

---

## 8. Forms & Validation

### 8.1 Form Libraries

- **React Hook Form** - Form state management (lightweight)
- **Zod** - Schema validation (TypeScript-first)

### 8.2 Example: Product Estimation Form

```typescript
// Using React Hook Form + Zod
const estimateSchema = z.object({
  productName: z.string().min(3).max(100),
  condition: z.enum(['new', 'like_new', 'good', 'fair', 'poor']),
  description: z.string().optional(),
  mode: z.enum(['fast', 'accurate', 'knowledge_based']).default('accurate'),
});

type EstimateForm = z.infer<typeof estimateSchema>;

export const EstimateForm = () => {
  const { register, handleSubmit, formState: { errors } } = useForm<EstimateForm>({
    resolver: zodResolver(estimateSchema),
  });

  const onSubmit = async (data: EstimateForm) => {
    // Call API with form data
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      {/* Form fields with error messages */}
    </form>
  );
};
```

---

## 9. Routing Structure

**React Router v7 Setup:**

```typescript
// routes.tsx
const routes = [
  {
    path: '/',
    element: <AppLayout />,
    children: [
      { path: '', element: <Dashboard /> },
      { path: 'upload', element: <Upload /> },
      { path: 'history', element: <History /> },
      { path: 'subscription', element: <Subscription /> },
      { path: 'settings', element: <Settings /> },
    ],
  },
  {
    path: '/auth',
    children: [
      { path: 'login', element: <Login /> },
      { path: 'register', element: <Register /> },
      { path: 'forgot-password', element: <ForgotPassword /> },
    ],
  },
  {
    path: '/admin',
    element: <ProtectedRoute requiredRole="admin"><AdminLayout /></ProtectedRoute>,
    children: [
      { path: 'dashboard', element: <AdminDashboard /> },
      { path: 'users', element: <AdminUsers /> },
      // ... more admin routes
    ],
  },
];
```

---

## 10. Mobile-First Responsive Design

### 10.1 Breakpoints (Tailwind)

- **Mobile:** < 640px (sm)
- **Tablet:** 640px - 1024px (md to lg)
- **Desktop:** > 1024px (xl+)

### 10.2 Mobile Optimizations

| Screen | Changes |
|--------|---------|
| **Mobile (<640px)** | Full-width, bottom nav, single column, touch-optimized buttons |
| **Tablet (640-1024px)** | Sidebar hidden by default, 2-column layouts, optimized spacing |
| **Desktop (>1024px)** | Sidebar visible, 3+ column layouts, hover states, full features |

**Example: Upload Page**
```typescript
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
  <PhotoUploadCard className="md:col-span-2" /> {/* Full width on mobile */}
  <SidePanel className="hidden md:block" /> {/* Hidden on mobile */}
</div>
```

---

## 11. Performance & Optimization

### 11.1 Code Splitting & Lazy Loading

```typescript
const AdminDashboard = React.lazy(() => import('./pages/AdminDashboard'));

const routes = [
  {
    path: '/admin',
    element: (
      <Suspense fallback={<Loading />}>
        <AdminDashboard />
      </Suspense>
    ),
  },
];
```

### 11.2 Image Optimization

- Compress uploaded images to <500KB
- Use WebP with PNG fallback
- Lazy load history thumbnails
- CDN delivery via S3 CloudFront

### 11.3 Caching Strategy

- **API responses:** Cache via React Query with 5-minute stale time
- **JWT tokens:** Store in memory (not localStorage to prevent XSS)
- **Images:** Browser cache (Cache-Control headers)

### 11.4 Bundle Analysis

- Target bundle size: <300KB (gzip)
- Use Webpack Bundle Analyzer in CI

---

## 12. Accessibility (WCAG 2.1 AA)

### 12.1 Keyboard Navigation

- Tab through all interactive elements
- Escape closes modals
- Enter submits forms
- Focus indicators visible

### 12.2 Screen Reader Support

- Semantic HTML (`<button>`, `<form>`, `<article>`)
- ARIA labels for icon-only buttons
- Form labels associated with inputs
- Announcement of state changes (e.g., "Analysis complete")

### 12.3 Color Contrast

- Minimum 4.5:1 contrast ratio for normal text
- 3:1 for large text
- No color-only information (use icons + text)

---

## 13. Localization (i18n)

**Setup: i18next + React Hook for translations**

```typescript
// locales/id.json
{
  "upload": {
    "title": "Unggah Foto Produk",
    "placeholder": "Nama produk, misal: iPhone 13 Pro Max",
    "buttonAnalyze": "Estimasi Harga",
  },
  "subscription": {
    "free": "Gratis",
    "premium": "Premium - Rp 65.000/bulan",
  }
}

// In component:
import { useTranslation } from 'react-i18next';

export const Upload = () => {
  const { t } = useTranslation();
  return <h1>{t('upload.title')}</h1>;
};
```

---

## 14. Error Handling & User Feedback

### 14.1 Error States

| Error | User Message | Action |
|-------|--------------|--------|
| Network error | "Koneksi terputus. Coba lagi?" | Retry button |
| Invalid file | "File terlalu besar. Max 10MB" | Clear input, try again |
| Quota exceeded | "Limit harian tercapai. Upgrade?" | Upgrade CTA |
| API error (500) | "Terjadi kesalahan. Hubungi support" | Show trace ID, contact form |

### 14.2 Toast Notifications

```typescript
import { useToast } from '@/components/ui/use-toast';

const { toast } = useToast();

// Success
toast({
  title: "Estimasi Selesai",
  description: "Harga diperkirakan Rp 5.3 juta",
  variant: "default",
});

// Error
toast({
  title: "Gagal Menganalisis",
  description: "Coba unggah foto lain",
  variant: "destructive",
});
```

---

## 15. Security Best Practices

### 15.1 Frontend Security

- **XSS Prevention:** Sanitize user inputs, use React's built-in escaping
- **CSRF Protection:** Token-based CSRF for state-changing operations
- **JWT Storage:** Keep tokens in memory/session storage (not localStorage)
- **HTTPS Only:** Enforce HTTPS, use secure cookie flags

### 15.2 Sensitive Data Handling

- Never log sensitive data (tokens, payment info)
- Mask credit card numbers in UI
- Validate input on client (but also on backend)

---

## 16. Testing Strategy

### 16.1 Test Coverage

| Layer | Tests | Coverage Target |
|-------|-------|-----------------|
| **Unit** | Component logic, hooks, utilities | >80% |
| **Integration** | Component + API interactions | >60% |
| **E2E** | Critical user flows (upload, payment) | >5 key flows |

### 16.2 Testing Tools

- **Jest** - Unit tests
- **React Testing Library** - Component tests
- **Cypress** - E2E tests

---

## 17. Deployment & Build

### 17.1 Build Process

```bash
npm run build  # Vite build → dist/
```

**Output:**
- index.html (entry point)
- assets/js/*.js (code-split bundles)
- assets/css/*.css (Tailwind + component styles)

### 17.2 Deployment Target

- **Hosting:** Vercel / Netlify (auto-deploy on git push)
- **Domain:** detectpricebyphoto.com / detect-price.id
- **CDN:** Vercel/Netlify CDN for static assets
- **Environment Variables:** API_BASE_URL, MIDTRANS_CLIENT_KEY

---

## 18. Analytics & Monitoring

### 18.1 User Behavior Tracking

- Page views, feature usage, conversion funnels
- Tool: Plausible Analytics (privacy-focused) or Mixpanel

### 18.2 Error Tracking

- JavaScript errors, API errors, network issues
- Tool: Sentry

### 18.3 Performance Monitoring

- Page load time, Time to Interactive (TTI), Cumulative Layout Shift (CLS)
- Tool: Web Vitals API + custom dashboard

---

## 19. Success Criteria & Launch Checklist

### 19.1 Must-Have (MVP Launch)

- [ ] Upload & estimation working end-to-end
- [ ] Authentication (login/register) functional
- [ ] Subscription management (Free → Premium upgrade)
- [ ] Payment integration with Midtrans
- [ ] History view showing past analyses
- [ ] Mobile responsive (tested on iOS/Android)
- [ ] Accessibility basics (keyboard nav, ARIA labels)
- [ ] <2 second API response times
- [ ] Error handling for all key flows

### 19.2 Nice-to-Have (Phase 2)

- [ ] Admin dashboard (user management, analytics)
- [ ] Batch upload feature
- [ ] Dark mode
- [ ] Advanced analytics for users
- [ ] API integration guide
- [ ] Referral program
- [ ] Social sharing (WhatsApp, Instagram)

---

## 20. Future Roadmap

**Phase 2 (3 months post-launch):**
- Batch upload UI
- Advanced analytics dashboard
- API documentation & developer tools
- Integration with marketplace APIs (Tokopedia, Shopee)

**Phase 3 (6 months post-launch):**
- Mobile app (iOS/Android) using React Native / Flutter
- Predictive pricing recommendations
- AI-powered bulk pricing suggestions
- B2B enterprise tier with custom pricing

---

## Appendix: Component Prop Types (TypeScript)

```typescript
// PhotoUploadCard
interface PhotoUploadCardProps {
  onPhotoSelect: (file: File, preview: string) => void;
  isLoading?: boolean;
  maxSizeKB?: number;
}

// PriceEstimateResult
interface PriceEstimateResultProps {
  analysis: {
    estimated_price_min: number;
    estimated_price_max: number;
    estimated_price_median: number;
    confidence_score: number;
    reasoning: string;
  };
  onShare?: (method: 'whatsapp' | 'sms' | 'copy') => void;
}

// SubscriptionCard
interface SubscriptionCardProps {
  currentPlan: 'free' | 'premium' | 'enterprise';
  renewalDate: string;
  onUpgrade?: () => void;
  onCancel?: () => void;
}
```

---

**Document Version:** 1.0  
**Last Updated:** 2025-11-08  
**Next Review:** 2025-12-08