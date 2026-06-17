# Skyvern UI Patterns

> **Source**: `Z:\10_WORKPLACE\Ti\Ti-learning-lab\01_Learning\lab\03_Knowledge\browser\skyvern\skyvern-frontend\`
> **Applied to**: Ti Router UI (`apps/router/ui/`)
> **Date**: 2026-05-04

## Overview

Skyvern là một AI automation platform với React UI rich features. Bài viết này ghi lại các UI patterns học được từ Skyvern và cách áp dụng vào Ti Router.

---

## 1. Credentials Management Pattern

### Location
- `src/routes/credentials/CredentialsPage.tsx`
- `src/routes/credentials/CredentialsList.tsx`
- `src/routes/credentials/CredentialItem.tsx`

### Pattern Components

#### 1.1 Page Structure
```tsx
// Tabs để phân loại credentials
<Tabs value={activeTab} onValueChange={handleTabChange}>
  <TabsList>
    <TabsTrigger value="passwords">Passwords</TabsTrigger>
    <TabsTrigger value="creditCards">Credit Cards</TabsTrigger>
    <TabsTrigger value="secrets">Secrets</TabsTrigger>
    <TabsTrigger value="twoFactor">2FA</TabsTrigger>
  </TabsList>
  <TabsContent value="passwords">
    <CredentialsList filter="password" />
  </TabsContent>
</Tabs>
```

**Key Points**:
- Sử dụng Tabs để phân loại content
- URL search params để sync tab state
- Filter prop để lọc items theo type

#### 1.2 List Component
```tsx
function CredentialsList({ filter, onStartBackgroundTest }: Props) {
  const [page, setPage] = useState(1);
  const { data: credentials, isLoading } = useCredentialsQuery({
    page,
    page_size: PAGE_SIZE,
  });

  // Loading skeleton
  if (isLoading) {
    return (
      <div className="space-y-5">
        <Skeleton className="h-20 w-full" />
        <Skeleton className="h-20 w-full" />
      </div>
    );
  }

  // Empty state
  if (filteredCredentials.length === 0 && page === 1) {
    return (
      <div className="rounded-md border border-slate-700 bg-slate-elevation1 p-6">
        {filter ? EMPTY_MESSAGE[filter] : "No credentials stored yet."}
      </div>
    );
  }

  // Pagination
  return (
    <div className="space-y-5">
      {filteredCredentials.map((credential) => (
        <CredentialItem key={credential.credential_id} credential={credential} />
      ))}
      <Pagination>
        <PaginationPrevious onClick={() => setPage((prev) => Math.max(1, prev - 1))} />
        <PaginationLink>{page}</PaginationLink>
        <PaginationNext onClick={() => setPage((prev) => prev + 1)} />
      </Pagination>
    </div>
  );
}
```

**Key Points**:
- Custom hook cho data fetching (`useCredentialsQuery`)
- Loading skeleton với Skeleton component
- Empty state messages
- Pagination với page state
- Filter logic client-side

#### 1.3 Item Component
```tsx
function CredentialItem({ credential, onStartBackgroundTest }: Props) {
  const [editModalOpen, setEditModalOpen] = useState(false);
  const activeTest = useCredentialTestStore((s) =>
    s.activeTest?.credentialId === credential.credential_id
      ? s.activeTest
      : null,
  );

  // Conditional rendering dựa trên type
  let credentialDetails = null;
  if (isPasswordCredential(credentialData)) {
    credentialDetails = (
      <div className="border-l pl-5">
        {/* Password details */}
      </div>
    );
  } else if (isCreditCardCredential(credentialData)) {
    credentialDetails = (
      <div className="border-l pl-5">
        {/* Credit card details */}
      </div>
    );
  }

  return (
    <div className="flex gap-5 rounded-lg bg-slate-elevation2 p-4">
      {/* Header với name, ID, status indicators */}
      <div className="w-48 space-y-2">
        <p className="truncate" title={credential.name}>{credential.name}</p>
        <p className="text-sm text-slate-400">{credential.credential_id}</p>
        {activeTest && (
          <div className="flex items-center gap-1 text-xs">
            <ReloadIcon className="animate-spin text-blue-400" />
            <span className="text-blue-400">Testing login</span>
          </div>
        )}
      </div>

      {/* Collapsible details với border-left */}
      {credentialDetails}

      {/* Action buttons */}
      <div className="ml-auto flex gap-1">
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <Button onClick={() => setEditModalOpen(true)}>
                <Pencil1Icon />
              </Button>
            </TooltipTrigger>
            <TooltipContent>Edit Credential</TooltipContent>
          </Tooltip>
        </TooltipProvider>
        <DeleteCredentialButton credential={credential} />
      </div>

      {/* Edit modal */}
      <CredentialsModal
        isOpen={editModalOpen}
        onOpenChange={setEditModalOpen}
        editingCredential={credential}
      />
    </div>
  );
}
```

**Key Points**:
- Conditional rendering dựa trên type
- Collapsible details với `border-l pl-5`
- Status indicators (active test, browser profile)
- Tooltip cho action buttons
- Modal cho edit action
- Stop propagation cho delete button

---

## 2. Task Details Pattern

### Location
- `src/routes/tasks/detail/TaskDetails.tsx`
- `src/components/SwitchBarNavigation.tsx`

### Pattern Components

#### 2.1 Multi-View Navigation
```tsx
// SwitchBarNavigation component
function SwitchBarNavigation({ options }: Props) {
  const [searchParams] = useSearchParams();

  return (
    <div className="flex w-fit gap-2 rounded-sm border border-slate-700 p-2">
      {options.map((option) => (
        <NavLink
          to={`${option.to}?${searchParams.toString()}`}
          replace
          key={option.to}
          className={({ isActive }) =>
            cn(
              "flex cursor-pointer items-center justify-center rounded-sm px-3 py-2",
              { "bg-slate-700": isActive }
            )
          }
        >
          {option.icon && <span className="mr-1">{option.icon}</span>}
          {option.label}
        </NavLink>
      ))}
    </div>
  );
}

// Usage
<SwitchBarNavigation
  options={[
    { label: "Actions", to: "actions" },
    { label: "Recording", to: "recording" },
    { label: "Parameters", to: "parameters" },
    { label: "Diagnostics", to: "diagnostics" },
  ]}
/>
<Outlet /> {/* Nested routes */}
```

**Key Points**:
- NavLink với `isActive` styling
- URL search params preservation
- Outlet cho nested routes
- Icon support

#### 2.2 Detail Page Structure
```tsx
function TaskDetails() {
  const taskId = useFirstParam("taskId", "runId");

  // Multiple queries với enabled conditions
  const { data: task, isLoading } = useTaskQuery({ id: taskId });
  const { data: workflowRun } = useQuery({
    queryKey: ["taskWorkflowRun", task?.workflow_run_id],
    queryFn: async () => { /* ... */ },
    enabled: !!task?.workflow_run_id,
  });
  const { data: workflow } = useQuery({
    queryKey: ["workflow", workflowRun?.workflow_id],
    queryFn: async () => { /* ... */ },
    enabled: !!workflowRun?.workflow_id,
  });

  // Mutation với invalidateQueries
  const cancelTaskMutation = useMutation({
    mutationFn: async () => { /* ... */ },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["task", taskId] });
      queryClient.invalidateQueries({ queryKey: ["tasks"] });
      if (task?.workflow_run_id) {
        queryClient.invalidateQueries({ queryKey: ["workflowRun", task.workflow_run_id] });
      }
    },
  });

  return (
    <div className="flex flex-col gap-8">
      {/* Header với actions */}
      <header>
        <div className="flex items-center justify-between">
          <span className="text-3xl">{taskId}</span>
          <StatusBadge status={task.status} />
        </div>
        <ApiWebhookActionsMenu getOptions={() => { /* ... */ }} />
      </header>

      {/* Conditional sections */}
      {task?.status === Status.Completed && (
        <div>
          <Label>Extracted Information</Label>
          <CodeEditor value={JSON.stringify(task.extracted_information)} readOnly />
        </div>
      )}

      {/* Navigation tabs */}
      <SwitchBarNavigation options={[...]} />
      <Outlet />
    </div>
  );
}
```

**Key Points**:
- Multiple queries với enabled conditions
- Mutation với invalidateQueries
- Conditional rendering dựa trên status
- CodeEditor cho JSON display
- StatusBadge component

---

## 3. Workflow Page Pattern

### Location
- `src/routes/workflows/WorkflowPage.tsx`

### Pattern Components

#### 3.1 Table với Expandable Rows
```tsx
function WorkflowPage() {
  const [expandedRows, setExpandedRows] = useState<Set<string>>(new Set());

  return (
    <Table>
      <TableBody>
        {workflowRuns?.map((workflowRun) => {
          const isExpanded = expandedRows.has(workflowRun.workflow_run_id);

          return (
            <React.Fragment key={workflowRun.workflow_run_id}>
              {/* Main row - click để navigate */}
              <TableRow
                onClick={() => navigate(`/runs/${workflowRun.workflow_run_id}`)}
                className="cursor-pointer"
              >
                <TableCell>{workflowRun.workflow_run_id}</TableCell>
                <TableCell><StatusBadge status={workflowRun.status} /></TableCell>
                <TableCell>{basicLocalTimeFormat(workflowRun.created_at)}</TableCell>
                <TableCell>
                  {/* Toggle button cho expand */}
                  <Button
                    onClick={(e) => {
                      e.stopPropagation();
                      toggleParametersExpanded(workflowRun.workflow_run_id);
                    }}
                  >
                    <MixerHorizontalIcon />
                  </Button>
                </TableCell>
              </TableRow>

              {/* Expanded parameters section */}
              {isExpanded && (
                <TableRow>
                  <TableCell colSpan={5} className="bg-slate-900/50">
                    <WorkflowRunParameters workflowRunId={workflowRun.workflow_run_id} />
                  </TableCell>
                </TableRow>
              )}
            </React.Fragment>
          );
        })}
      </TableBody>
    </Table>
  );
}
```

**Key Points**:
- React.Fragment cho multiple rows per item
- Stop propagation cho toggle button
- Ctrl+click để mở trong tab mới
- Background color cho expanded row
- Custom hook cho expand state (`useParameterExpansion`)

#### 3.2 Search với Debounce
```tsx
function WorkflowPage() {
  const [search, setSearch] = useState("");
  const [debouncedSearch] = useDebounce(search, 500);

  const { data: workflowRuns } = useWorkflowRunsQuery({
    search: debouncedSearch,
  });

  return (
    <TableSearchInput
      value={search}
      onChange={(value) => {
        setSearch(value);
        const params = new URLSearchParams(searchParams);
        params.set("page", "1");
        setSearchParams(params, { replace: true });
      }}
    />
  );
}
```

**Key Points**:
- Debounce với `use-debounce` library
- Reset page khi search changes
- URL sync với search params

#### 3.3 Status Filter Dropdown
```tsx
function WorkflowPage() {
  const [statusFilters, setStatusFilters] = useState<Array<Status>>([]);

  return (
    <StatusFilterDropdown
      values={statusFilters}
      onChange={setStatusFilters}
    />
  );
}
```

**Key Points**:
- Array filter cho multiple selection
- Custom component cho status filter

---

## 4. Custom Hooks Pattern

### 4.1 useCredentialModalState
```tsx
function useCredentialModalState() {
  const [isOpen, setIsOpen] = useState(false);
  const [type, setType] = useState<CredentialModalTypes>(CredentialModalTypes.PASSWORD);

  return {
    isOpen,
    setIsOpen,
    type,
    setType,
  };
}
```

**Key Points**:
- Encapsulate modal state
- Type-safe với enum

### 4.2 useBackgroundCredentialTest
```tsx
function useBackgroundCredentialTest() {
  const startBackgroundTest = (credentialId: string, url: string) => {
    // Start workflow run in background
  };

  return { startBackgroundTest };
}
```

**Key Points**:
- Encapsulate complex logic
- Return functions for actions

### 4.3 useCredentialTestStore (Zustand)
```tsx
const useCredentialTestStore = create((set) => ({
  activeTest: null,
  setActiveTest: (test) => set({ activeTest: test }),
}));

// Usage
const activeTest = useCredentialTestStore((s) =>
  s.activeTest?.credentialId === credential.credential_id
    ? s.activeTest
    : null,
);
```

**Key Points**:
- Zustand cho global state
- Selector pattern cho efficiency

---

## 5. Dialog/Modal Pattern

### Pattern
```tsx
function MyDialog({ isOpen, onClose, data }: Props) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (isOpen && data) {
      loadData();
    }
  }, [isOpen, data]);

  const loadData = async () => {
    setLoading(true);
    setError(null);
    try {
      // Fetch data
    } catch (err) {
      setError('Failed to load data');
    } finally {
      setLoading(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="overlay" onClick={onClose}>
      <div className="dialog" onClick={(e) => e.stopPropagation()}>
        {/* Header */}
        <div className="header">
          <h2>Title</h2>
          <button onClick={onClose}><XIcon /></button>
        </div>

        {/* Body */}
        <div className="body">
          {loading ? <Spinner /> : error ? <Error /> : <Content />}
        </div>
      </div>
    </div>
  );
}
```

**Key Points**:
- Overlay click để close
- Stop propagation cho dialog content
- useEffect để load data khi open
- Loading/error/content states

---

## 6. Applied to Ti Router

### Implementation: Auth Files with Models View

#### File Structure
```
apps/router/ui/
├── src/
│   ├── components/
│   │   └── AuthFileModelsDialog.tsx          # New
│   │   └── AuthFileModelsDialog.module.scss # New
│   ├── pages/
│   │   └── AuthFilesPage.tsx                # Modified
│   │   └── AuthFilesPage.module.scss       # Modified
│   └── services/
│       └── api.ts                           # Modified
```

#### Changes Made

1. **API Service** (`api.ts`)
   ```typescript
   async getProviderModels(providerName: string): Promise<{ models: string[] }> {
     const providers = await this.getProviders();
     const provider = providers.providers.find(p => p.name === providerName);
     if (!provider) {
       return { models: [] };
     }
     return { models: provider.models };
   }
   ```

2. **Dialog Component** (`AuthFileModelsDialog.tsx`)
   - Loading state với spinner
   - Error handling cho provider không tìm thấy
   - Empty state khi không có models
   - Grid layout cho model cards
   - Backdrop blur effect

3. **Page Component** (`AuthFilesPage.tsx`)
   - Click handler trên file card
   - Stop propagation cho delete button
   - "View models" footer với Cpu icon
   - Dialog integration

4. **Styling** (`AuthFilesPage.module.scss`)
   - Hover effect: border color + background gradient
   - Footer styling
   - Cursor pointer cho clickable cards

---

## 7. Key Takeaways

### Best Practices
1. **Custom Hooks**: Encapsulate state và logic
2. **Loading States**: Luôn có skeleton hoặc spinner
3. **Empty States**: Clear messages khi không có data
4. **Error Handling**: User-friendly error messages
5. **Stop Propagation**: Cho nested click handlers
6. **URL Sync**: State sync với URL params
7. **Debounce**: Cho search inputs
8. **Conditional Rendering**: Dựa trên status/type
9. **Tooltip**: Cho action buttons
10. **Modal/Dialog**: Với backdrop blur và animations

### Component Patterns
1. **List-Item-Detail**: List → Item → Detail hierarchy
2. **Expandable Rows**: Table với collapsible content
3. **Multi-View Navigation**: Tabs với nested routes
4. **Card Grid**: Responsive grid layout
5. **Status Indicators**: Visual feedback cho async operations

### Styling Patterns
1. **Hover Effects**: Border color + background gradient
2. **Transitions**: Smooth transitions cho interactions
3. **Backdrop Blur**: Modal overlay effect
4. **Border Left**: Visual separator cho details
5. **Color Coding**: Status-based colors

---

## 8. Future Enhancements

### Potential Improvements for Ti Router
1. **Add Tabs** cho Auth Files (bym type: json, oauth, token)
2. **Add Search** với debounce
3. **Add Status Filter** cho providers
4. **Add Expandable Details** cho file content preview
5. **Add Bulk Actions** cho delete multiple files
6. **Add Drag & Drop** cho file upload
7. **Add Real-time Updates** với WebSocket
8. **Add Code Editor** cho viewing auth file content

---

## 9. References

- **Skyvern Frontend**: `Z:\10_WORKPLACE\Ti\Ti-learning-lab\01_Learning\lab\03_Knowledge\browser\skyvern\skyvern-frontend\`
- **Ti Router UI**: `Z:\10_WORKPLACE\Ti\apps\router\ui\`
- **Radix UI**: Component library used by Skyvern
- **TanStack Query**: Data fetching library
- **Zustand**: State management library
