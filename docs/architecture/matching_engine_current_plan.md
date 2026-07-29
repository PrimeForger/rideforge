MatchingEngine
│
▼
CandidateSearcher
│
▼
H3Strategy
│
▼
RingExpander
│
▼
AdaptiveRingExpansionPolicy
│
├───────────────┐
▼ --------------▼
DriverDensityProvider ExpansionStrategy
│--------------------- ▲
▼ ---------------------│
DensityClassifier ─────┘
│
▼
RingDecision

---

---

------Currrent OwnerShip------

MatchingEngine
│
▼
Search.Request
│
▼
H3Strategy
│
▼
PolicyInput
│
▼
BudgetFactory
│
▼
SearchBudget
│
▼
AdaptiveRingExpansionPolicy

......Note......
PolicyInput stays narrowly focused on what the budget policy needs. As later stages introduce retry-aware budgets, region-specific policies, or premium ride behavior, you can extend PolicyInput without coupling those concerns to the search request or runtime search state

---

- Discovery pipeline now looks roughly like this:

- ***

  MatchingEngine
  │
  ▼
  CandidateSearcher
  │
  ▼
  H3Strategy
  │
  ├── SearchBudgetFactory
  │
  ├── RingExpander
  │
  └── H3Service
  │
  ▼
  AdaptiveRingExpansionPolicy
  │
  ├── DriverDensityProvider
  ├── DensityClassifier
  ├── ExpansionProfileProvider
  └── ExpansionStrategy

- ***
