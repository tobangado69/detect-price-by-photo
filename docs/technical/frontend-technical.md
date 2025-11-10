# Frontend Technical Notes

## 1. Repository Layout

```
apps/frontend/
├── public/                 # static assets (favicons, logos)
├── src/
│   ├── App.tsx             # route composition
│   ├── main.tsx            # ReactDOM bootstrap
│   ├── components/
│   │   ├── CameraCaptureModal/   # live camera capture UI
│   │   ├── PhotoPreview/
│   │   ├── Layout/               # RootLayout, Sidebar, Header
│   │   └── Admin/ModelConfigPanel
│   ├── pages/
│   │   ├── Home.tsx
│   │   ├── Dashboard.tsx
│   │   ├── Upload.tsx
│   │   ├── Pricing.tsx
│   │   └── Admin.tsx
│   ├── hooks/              # useCamera, useUpload
│   ├── lib/                # apiClient, formatters
│   ├── store/              # Zustand stores (auth, subscription, analyses, admin)
│   ├── styles/             # global.css + theme overrides
│   └── router.tsx          # route definitions (optional)
├── package.json
└── vite.config.ts
```

## 2. Camera Capture Flow

- Use the browser MediaDevices API (`navigator.mediaDevices.getUserMedia`) to request `video` stream.
- Provide a fallback `<input type="file" accept="image/*" capture="environment">` for devices that block live capture.
- Convert the captured frame to a `Blob` and upload via `FormData`.

```tsx
import { useEffect, useRef, useState } from 'react';

type Props = {
  open: boolean;
  onClose: () => void;
  onCapture: (file: File, preview: string) => void;
};

export const CameraCaptureModal: React.FC<Props> = ({ open, onClose, onCapture }) => {
  const videoRef = useRef<HTMLVideoElement>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [stream, setStream] = useState<MediaStream | null>(null);

  useEffect(() => {
    if (!open) return;
    navigator.mediaDevices.getUserMedia({ video: { facingMode: 'environment' } })
      .then((s) => {
        setStream(s);
        if (videoRef.current) {
          videoRef.current.srcObject = s;
          videoRef.current.play();
        }
      })
      .catch(() => {
        // fallback UI: instruct user to use file upload
      });
    return () => {
      s?.getTracks().forEach((track) => track.stop());
    };
  }, [open]);

  const handleCapture = () => {
    if (!canvasRef.current || !videoRef.current) return;
    const ctx = canvasRef.current.getContext('2d');
    if (!ctx) return;
    canvasRef.current.width = videoRef.current.videoWidth;
    canvasRef.current.height = videoRef.current.videoHeight;
    ctx.drawImage(videoRef.current, 0, 0);

    canvasRef.current.toBlob((blob) => {
      if (!blob) return;
      const file = new File([blob], `capture-${Date.now()}.jpg`, { type: blob.type });
      const preview = canvasRef.current!.toDataURL('image/jpeg', 0.92);
      onCapture(file, preview);
      onClose();
    }, 'image/jpeg', 0.92);
  };

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-sm">
        <video ref={videoRef} playsInline className="w-full rounded-md bg-black" />
        <canvas ref={canvasRef} className="hidden" />
        <div className="mt-4 flex justify-between">
          <Button variant="secondary" onClick={onClose}>Cancel</Button>
          <Button onClick={handleCapture}>Capture</Button>
        </div>
      </DialogContent>
    </Dialog>
  );
};
```

When the modal closes with a captured file, the `Upload` page calls:

```tsx
const formData = new FormData();
formData.append('photo', file);
formData.append('product_name', productName);
formData.append('condition', condition);
await apiClient.post('/analyses/upload', formData, {
  headers: { 'Content-Type': 'multipart/form-data' },
});
```

## 3. Upload Page Integration

1. User opens `CameraCaptureModal` or drags a file onto `PhotoUploadCard`.
2. `PhotoPreview` displays the captured image. Store the `File` and preview in local state.
3. On submission, dispatch `AnalysesStore.createAnalysis(file, payload)` which performs the upload and polls `/api/v1/analyses/{id}` for results.
4. On success, push the analysis to `AnalysesStore.recentAnalyses`.

Zustand slice (simplified):

```ts
type AnalysesState = {
  recentAnalyses: AnalysisSummary[];
  createAnalysis: (file: File, input: EstimateInput) => Promise<void>;
};

export const useAnalysesStore = create<AnalysesState>((set, get) => ({
  recentAnalyses: [],
  async createAnalysis(file, input) {
    const formData = new FormData();
    formData.append('photo', file);
    Object.entries(input).forEach(([key, value]) => formData.append(key, value));

    const { data } = await apiClient.post('/analyses/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });

    set((state) => ({
      recentAnalyses: [data.analysis, ...state.recentAnalyses].slice(0, 20),
    }));
  },
}));
```

## 4. Admin – ModelConfigPanel

Purpose: allow admins to review allowed models, set the global default, and adjust routing. The panel consumes the admin endpoints:

- `GET /api/v1/admin/models`
- `PUT /api/v1/admin/models/default`
- `PUT /api/v1/admin/models/:model_id`

UI blueprint:

```
┌───────────────────────────────────────────────┐
│ Default Model: [ GPT-4o mini ▼ ]  (Save)      │
├───────────────────────────────────────────────┤
│ Model Catalog                                 │
│ ┌───────────────────────────────────────────┐ │
│ │ GPT-4o mini     cost $0.002 / 1k tokens   │ │
│ │ Modes: fast, accurate                     │ │
│ │ Fallback: Claude 3.5 Haiku                │ │
│ │ [Toggle Active] [Edit Modes]              │ │
│ └───────────────────────────────────────────┘ │
│ ...                                           │
└───────────────────────────────────────────────┘
```

Implementation sketch:

```tsx
const ModelConfigPanel: React.FC = () => {
  const { models, fetchCatalog, setDefault, updateModel } = useAdminStore();

  useEffect(() => {
    fetchCatalog();
  }, [fetchCatalog]);

  return (
    <Card>
      <Select value={models.defaultId} onValueChange={(id) => setDefault(id)}>
        {models.items.map((model) => (
          <SelectItem key={model.id} value={model.id}>
            {model.displayName} • {currency(model.costPer1kTokens)}
          </SelectItem>
        ))}
      </Select>
      <Button onClick={() => setDefault(models.pendingDefault)}>Save</Button>

      <div className="mt-6 space-y-4">
        {models.items.map((model) => (
          <ModelCard key={model.id} model={model} onUpdate={updateModel} />
        ))}
      </div>
    </Card>
  );
};
```

`useAdminStore` interacts with the backend, handles optimistic updates, and surfaces errors via `toast`.

> The customer-facing upload flow does **not** present a model picker. Any UI that references AI models must live behind admin routes guarded by role checks.

## 5. API Client Helpers

Extend the Axios client with form-data support:

```ts
export const uploadAnalysis = (formData: FormData) =>
  apiClient.post('/analyses/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });

export const listModels = () => apiClient.get<AIModel[]>('/admin/models');
export const setDefaultModel = (modelId: string) =>
  apiClient.put('/admin/models/default', { model_id: modelId });
export const updateModelConfig = (modelId: string, body: Partial<AIModel>) =>
  apiClient.put(`/admin/models/${modelId}`, body);
```

Centralize types in `packages/shared-types` so backend and frontend share DTOs.

## 6. Icon System

- Install and use [`@iconify/react`](https://iconify.design/docs/icon-components/react/) for all icons:
  ```bash
  pnpm add @iconify/react
  ```
- Wrap icons in a small helper module (`src/components/icons.tsx`) to avoid scattered imports.
- Example usage:
  ```tsx
  import { Icon } from '@iconify/react';

  <Icon icon="solar:camera-bold" className="h-6 w-6 text-primary" />
  ```
- Adopt consistent naming (`icon="mdi:upload"` etc.) and avoid mixing with other icon libraries to keep bundle size predictable.

## 7. Testing Notes

- Camera capture hook: stub `navigator.mediaDevices.getUserMedia` in Vitest.
- Upload flow: use [`msw`](https://mswjs.io/) to mock multipart endpoints.
- Admin model panel: verify PUT requests fire with correct payload and optimistic UI updates.

--- 

**Document Version:** 1.1  
**Last Updated:** 2025-11-08  
**Next Review:** 2025-12-08