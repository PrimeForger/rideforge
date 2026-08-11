# ADR-0016: AI-Assisted Dispatch Strategy

> **Status:** Accepted  
> **Date:** 2026-08-09  
> **Decision Type:** AI / Dispatch Architecture  
> **Scope:** Optional AI assistance for RideForge driver-dispatch decisions  
> **Owner:** RideForge Architecture / Engineering  
> **Supersedes:** None  
> **Superseded By:** None

---

# 1. Context

RideForge supports multiple dispatch strategies:

```text
Stand Dispatch
Smart Dispatch
AI-Assisted Dispatch
```

ADR-0015 establishes that Stand Dispatch and Smart Dispatch are first-class dispatch strategies and that AI-assisted dispatch is an optional enhancement.

The purpose of AI-assisted dispatch is not to replace the dispatch domain or its hard business rules.

Instead, AI/ML should help the system make better decisions from complex operational signals such as:

```text
Driver Location
Pickup ETA
Driver Availability
Demand
Supply
Historical Assignment Behaviour
Traffic
Time of Day
Day of Week
Zone Characteristics
Ride Characteristics
Driver State
Fairness Signals
```

A machine-learning model may identify relationships that are difficult to capture with a simple deterministic ranking formula.

However, introducing AI directly into a critical dispatch path also creates risks:

```text
Inference Latency
Model Failure
Feature Failure
Data Drift
Model Drift
Unexpected Predictions
Operational Cost
Versioning Complexity
Explainability Requirements
Safety / Fairness Concerns
```

The architecture therefore needs a controlled way to use AI without making the core dispatch system dependent on it.

---

# 2. Problem

RideForge must determine:

```text
Where AI should participate in dispatch
What AI is allowed to decide
What AI is not allowed to decide
How candidates are prepared
How AI scores candidates
How model failures are handled
How models are deployed
How model versions are tracked
How fallback works
How AI decisions are observed
How AI is evaluated before production use
```

The system must preserve:

```text
Deterministic Safety
Operational Control
Legal Compliance
Low Latency
Driver Fairness
Rider Experience
Recoverability
Model Governance
```

---

# 3. Decision

RideForge will use AI as an **optional intelligence layer inside the Smart Dispatch pipeline**.

The core architecture is:

```text
Ride Request
     │
     ▼
Ride Validation
     │
     ▼
Regional / Legal Validation
     │
     ▼
Dispatch Policy
     │
     ▼
Candidate Discovery
     │
     ▼
Hard Eligibility Filtering
     │
     ▼
Feature Construction
     │
     ▼
AI Ranking / Scoring
     │
     ▼
Policy Constraints
     │
     ▼
Driver Reservation
     │
     ▼
Assignment
```

The critical principle is:

> **AI may rank eligible candidates, but AI must not override hard safety, legal, operational, or state constraints.**

---

# 4. AI Is Not the Source of Truth

The AI model does not own:

```text
Ride State
Driver State
Legal Rules
Regional Policy
Assignment State
Payment State
```

Those remain owned by the appropriate domain services.

AI produces a decision-support output such as:

```text
candidate score
ranking
probability
prediction
```

The dispatch system remains responsible for the final assignment decision.

---

# 5. AI as a Bounded Decision Component

The model should be treated as:

```text
Input
 ↓
Inference
 ↓
Prediction / Ranking
 ↓
Dispatch Policy
```

not:

```text
Input
 ↓
AI
 ↓
Unrestricted Business Action
```

The AI component therefore operates inside explicit application boundaries.

---

# 6. Hard Constraints Come First

The dispatch pipeline must enforce hard constraints before invoking AI.

Conceptually:

```text
All Drivers
    ↓
Availability
    ↓
Vehicle Compatibility
    ↓
Regional Eligibility
    ↓
Legal Constraints
    ↓
Ride Compatibility
    ↓
Candidate Set
    ↓
AI Ranking
```

AI must never be used to decide whether an otherwise prohibited driver becomes eligible.

---

# 7. Soft Signals

AI may optimize soft signals such as:

```text
Pickup ETA
Expected Assignment Success
Expected Acceptance
Expected Driver Utilization
Expected Cancellation
Expected Rider Experience
Fairness Signals
Demand / Supply Conditions
```

These are optimization inputs, not permission to violate hard constraints.

---

# 8. Initial AI Responsibility

The initial AI responsibility should be deliberately narrow:

> **Rank an already eligible candidate set for Smart Dispatch.**

This avoids introducing unnecessary AI complexity into:

```text
Candidate Discovery
Driver Eligibility
Legal Validation
Assignment Transactions
```

---

# 9. AI Ranking Flow

The preferred flow is:

```text
Eligible Candidates
        │
        ▼
Feature Extraction
        │
        ▼
Feature Validation
        │
        ▼
Model Inference
        │
        ▼
Candidate Scores
        │
        ▼
Ranking
        │
        ▼
Policy / Safety Checks
        │
        ▼
Assignment Candidate
```

---

# 10. AI Must Remain Optional

The Smart Dispatch system must remain operational when AI is unavailable.

Preferred fallback:

```text
AI-Assisted Dispatch
        │
        ├── Success → AI Ranking
        │
        └── Failure → Deterministic Smart Ranking
```

The fallback should not require architectural reconfiguration during an incident.

---

# 11. Deterministic Fallback

The deterministic fallback should be a known, tested ranking strategy.

For example:

```text
Eligibility
    ↓
Pickup ETA
    ↓
Distance
    ↓
Operational Priority
```

The exact ranking formula is governed by Smart Dispatch implementation.

---

# 12. Why AI Is Optional

Making AI optional provides:

```text
Higher Availability
Predictable Failure Behaviour
Lower Operational Risk
Controlled Cost
Simpler Rollback
Easier Testing
```

AI should improve dispatch quality rather than become a new single point of failure.

---

# 13. AI in the Critical Path

AI may be used synchronously when:

```text
Inference Latency
+
Infrastructure Reliability
+
Model Availability
```

fit within the dispatch latency budget.

The AI request must always have a strict timeout.

---

# 14. AI Timeout

A model request must never block dispatch indefinitely.

Conceptually:

```text
Candidate Set
     ↓
AI Inference
     ↓
Timeout
     ↓
Deterministic Smart Ranking
```

The timeout must be shorter than the total dispatch decision deadline.

---

# 15. AI Failure

AI failures include:

```text
Inference Timeout
Model Unavailable
Model Error
Feature Service Failure
Invalid Model Output
Serialization Error
Network Failure
Resource Exhaustion
```

All must have defined behaviour.

---

# 16. AI Failure Policy

The default policy is:

```text
AI Failure
    ↓
Record Failure
    ↓
Use Deterministic Smart Ranking
    ↓
Continue Dispatch
```

unless the dispatch policy explicitly defines another safe behaviour.

---

# 17. AI Failure Must Be Observable

AI failures must not silently disappear.

Record appropriate metadata such as:

```text
model_version
strategy
failure_code
latency
fallback_used
correlation_id
```

Avoid logging sensitive model inputs unnecessarily.

---

# 18. AI Model Output

A model may return:

```text
driver_id
score
model_version
inference_time
```

or a ranking over candidate IDs.

The model output should not directly mutate business state.

---

# 19. Model Output Validation

Every model response must be validated.

Check:

```text
Candidate Exists
Candidate Is In Current Set
Score Is Valid
Required Fields Exist
Model Version Is Known
Response Is Within Expected Bounds
```

Invalid output must trigger fallback.

---

# 20. AI Cannot Select an Ineligible Driver

If the model returns:

```text
Driver X
```

but Driver X is no longer eligible:

```text
Reject Model Result
```

and continue using the safe deterministic path.

---

# 21. Candidate Snapshot

The model should operate on a consistent candidate snapshot.

Conceptually:

```text
Candidate Set
+
Feature Snapshot
+
Policy Version
+
Model Version
```

This makes decisions more reproducible.

---

# 22. Revalidation Before Assignment

Even after AI ranking:

```text
Selected Candidate
        ↓
Revalidate Current State
        ↓
Reserve
        ↓
Assign
```

This is mandatory because driver state can change during inference.

---

# 23. Race Condition

Example:

```text
10:00:00
Driver A = Available

10:00:01
AI selects Driver A

10:00:02
Driver A accepts another ride

10:00:03
Assignment attempts Driver A
```

The assignment must reject the stale decision and continue safely.

---

# 24. AI Decision Does Not Reserve a Driver

Model inference should normally be treated as advisory.

The actual reservation happens after:

```text
Ranking
+
Current-State Validation
```

This prevents model latency from holding driver resources unnecessarily.

---

# 25. AI Feature Sources

Potential features include:

```text
Driver Location
Pickup Location
Destination
Pickup ETA
Distance
Traffic
Driver Availability
Driver Idle Time
Historical Acceptance
Historical Cancellation
Demand
Supply
Zone
Time of Day
Day of Week
Vehicle Type
Ride Type
```

Only approved and governed features should be used.

---

# 26. Feature Freshness

Features used for real-time dispatch must have defined freshness requirements.

For example:

```text
Driver Location
→ Very Fresh

Demand Estimate
→ Recent

Historical Statistics
→ Slower Changing
```

The exact freshness requirements belong to the AI feature architecture.

---

# 27. Missing Features

If a required feature is missing:

```text
Model Inference
```

may be unsafe.

The system should follow an explicit policy:

```text
Use Default
Use Alternate Feature
Exclude Candidate
Fallback to Deterministic Ranking
```

The default behaviour should be defined per feature/model.

---

# 28. Feature Validation

Before inference:

```text
Validate Feature Schema
Validate Types
Validate Ranges
Validate Freshness
Validate Required Features
```

Invalid feature sets should not be blindly sent to the model.

---

# 29. Feature Leakage

Training and serving features must avoid future information that would not be available at dispatch time.

For example:

```text
Future Driver Acceptance
Future Ride Completion
Future Cancellation
```

must not leak into real-time decision features.

---

# 30. Training / Serving Consistency

The same feature definitions should be used consistently between:

```text
Offline Training
Offline Evaluation
Online Inference
```

where practical.

Differences between training and serving features can cause model degradation.

---

# 31. Model Objective

The initial AI model should optimize a clearly defined dispatch objective.

Potential objectives include:

```text
Minimize Pickup ETA
Maximize Assignment Probability
Reduce Cancellation
Improve Driver Utilization
Improve Rider Experience
Improve System Efficiency
```

The objective must be measurable.

---

# 32. Multi-Objective Optimization

Dispatch is inherently multi-objective.

A model may need to balance:

```text
ETA
Acceptance
Fairness
Utilization
Cancellation
Cost
```

The final objective should not be defined solely by a single metric unless the business explicitly chooses that trade-off.

---

# 33. AI vs Business Policy

Business policy must remain outside the learned model where practical.

For example:

```text
Policy:
Driver must be legally eligible.

AI:
Among eligible drivers, estimate the best candidate.
```

This makes policy enforcement deterministic and auditable.

---

# 34. AI vs Fairness Policy

Fairness constraints should not depend entirely on the model learning them automatically.

Where required:

```text
Fairness Rules
```

should be enforced explicitly.

AI can optimize within those constraints.

---

# 35. AI Model Versioning

Every AI dispatch decision should be associated with a model version.

Example:

```text
model_version = dispatch-ranker-v3
```

This allows:

```text
Debugging
Experiment Analysis
Rollback
Performance Comparison
```

---

# 36. Policy Versioning

Model version alone is insufficient.

Also record:

```text
policy_version
```

because the same model may behave differently under different dispatch policies.

---

# 37. Feature Versioning

Where important, also track:

```text
feature_set_version
```

This provides a reproducible decision context:

```text
Policy Version
+
Feature Version
+
Model Version
```

---

# 38. Decision Metadata

A dispatch decision may retain:

```text
ride_id
dispatch_attempt_id
strategy
policy_version
model_version
feature_version
selected_driver_id
decision_timestamp
fallback_used
```

Retention should follow data governance requirements.

---

# 39. Model Registry

AI models should be managed through a controlled model registry.

The registry should identify:

```text
Model Name
Model Version
Training Dataset
Feature Version
Metrics
Status
Deployment Stage
Approval State
```

---

# 40. Model Lifecycle

The model lifecycle is:

```text
Experiment
   ↓
Training
   ↓
Evaluation
   ↓
Validation
   ↓
Registry
   ↓
Shadow
   ↓
Canary
   ↓
Production
   ↓
Monitoring
   ↓
Rollback / Retirement
```

---

# 41. Model Approval

A model must not become production-active merely because it was trained successfully.

Production promotion should require:

```text
Offline Evaluation
Safety Checks
Latency Validation
Fairness Review
Operational Approval
```

where applicable.

---

# 42. Shadow Mode

Before controlling real assignments, a new model can run in shadow mode.

```text
Production Dispatch
       │
       ├── Existing Strategy → Real Decision
       │
       └── New AI Model → Shadow Prediction
```

The shadow result is evaluated but does not control the ride.

---

# 43. Shadow Metrics

Compare:

```text
AI Ranking
vs
Current Ranking
```

using:

```text
Pickup ETA
Assignment Outcome
Acceptance
Cancellation
Fairness
Latency
```

---

# 44. Canary Deployment

After shadow validation:

```text
Small Traffic
      ↓
Observe
      ↓
Increase
```

The model should not immediately receive 100% of production dispatch traffic.

---

# 45. Rollback

Rollback must be simple.

Example:

```text
AI Model v4
     ↓
Problem
     ↓
AI Model v3
```

or:

```text
AI-Assisted
     ↓
Deterministic Smart
```

The second fallback must remain operational.

---

# 46. Rollback Trigger

Possible triggers include:

```text
Latency Regression
Assignment Failure Increase
Cancellation Increase
Fallback Rate Increase
Model Error Rate
Fairness Regression
Cost Increase
Operational Incident
```

---

# 47. AI Latency Budget

AI inference must fit within a strict dispatch latency budget.

Conceptually:

```text
Dispatch Budget
├── Candidate Discovery
├── Feature Retrieval
├── AI Inference
├── Revalidation
└── Assignment
```

AI should not consume the entire budget.

---

# 48. AI Cost Budget

AI inference may have direct infrastructure or model-serving cost.

The platform should track:

```text
Inference Count
Inference Cost
Cost per Dispatch
Fallback Rate
Cost by Region
Cost by Model Version
```

---

# 49. Candidate Limit

Do not send an unnecessarily large candidate set to the model.

Preferred:

```text
All Drivers
   ↓
Hard Filtering
   ↓
Top N Candidate Set
   ↓
AI Ranking
```

This reduces:

```text
Latency
Cost
Feature Computation
Model Complexity
```

---

# 50. Hierarchical Ranking

A useful architecture is:

```text
Stage 1
Cheap Deterministic Filtering

Stage 2
Deterministic Pre-Ranking

Stage 3
AI Ranking

Stage 4
Final Policy Constraints

Stage 5
Assignment
```

This should be used only if it provides measurable value.

---

# 51. AI Does Not Need to Score Every Driver

The model should normally operate on a bounded candidate set.

This keeps inference:

```text
Fast
Predictable
Cost-Controlled
```

---

# 52. Real-Time vs Batch AI

AI workloads should be separated into:

```text
Real-Time Inference
```

and:

```text
Offline Training / Batch Processing
```

Real-time dispatch must not depend on a slow training pipeline.

---

# 53. Training Isolation

Training workloads should not consume production dispatch resources in a way that degrades real-time inference.

Prefer separate:

```text
Compute
Queues
Workers
Resource Limits
```

where required.

---

# 54. Model Serving

The model-serving layer should provide:

```text
Low-Latency Inference
Version Selection
Health Checks
Metrics
Timeout Handling
Controlled Rollout
```

---

# 55. Model Serving Boundary

Dispatch should communicate with an abstract AI inference interface.

Conceptually:

```text
Dispatch Service
       ↓
AI Ranking Interface
       ↓
Model Serving
       ↓
Model
```

This avoids coupling dispatch logic to a specific ML framework.

---

# 56. Model Framework Independence

The dispatch service should not depend directly on:

```text
PyTorch
TensorFlow
XGBoost
LightGBM
ONNX Runtime
```

unless that dependency belongs inside the model-serving layer.

---

# 57. AI Service Technology

The model-serving implementation may use a technology appropriate for the model.

RideForge may use:

```text
Python / FastAPI
```

for AI services where appropriate, while keeping the core dispatch service in Go.

The architecture should treat this as an implementation boundary, not a domain dependency.

---

# 58. Go Dispatch Service

The core dispatch orchestration should remain compatible with the Go service architecture.

Conceptually:

```text
Go Dispatch
     ↓
Candidate Selection
     ↓
AI Inference Client
     ↓
AI Service
```

The AI service returns a structured prediction.

---

# 59. AI Service Failure Isolation

The AI service must not be allowed to:

```text
Block Goroutines Indefinitely
Exhaust Connection Pools
Consume Unbounded Memory
Trigger Retry Storms
```

Use:

```text
Timeout
Bounded Concurrency
Connection Limits
Circuit Breaking
```

where appropriate.

---

# 60. AI Concurrency

AI inference should have bounded concurrency.

Conceptually:

```text
Dispatch Requests
       ↓
Inference Pool
       ↓
Model Server
```

This prevents traffic spikes from overwhelming the model-serving layer.

---

# 61. AI Backpressure

When AI capacity is exhausted:

```text
AI Queue Full
      ↓
Fallback to Smart Dispatch
```

rather than allowing unlimited waiting.

---

# 62. AI Circuit Breaker

If model inference repeatedly fails:

```text
AI Calls
   ↓
Repeated Failures
   ↓
Circuit Open
   ↓
Deterministic Smart Dispatch
```

After a recovery period, limited AI calls may resume.

---

# 63. AI Health Checks

Model serving should expose health/readiness signals.

The dispatch service should know whether the model service is:

```text
Healthy
Degraded
Unavailable
```

Health checks must not replace actual inference error handling.

---

# 64. Model Warm-Up

Cold-start latency should be considered.

A production model should be warmed or provisioned appropriately when low latency is required.

---

# 65. AI Output Confidence

If the model provides confidence or probability:

```text
confidence
```

may be used as an operational signal.

However, confidence must not automatically be interpreted as:

```text
Correctness
```

unless the model has been properly calibrated and evaluated.

---

# 66. Low-Confidence Fallback

A model may define a safe confidence threshold:

```text
High Confidence
→ AI Decision

Low Confidence
→ Deterministic Ranking
```

This should be introduced only after evaluation demonstrates that it improves reliability.

---

# 67. Model Calibration

If probabilities are used operationally, calibration should be evaluated.

For example:

```text
Predicted 80%
```

should have a meaningful relationship to:

```text
Observed Outcome
```

if the probability is used as a business signal.

---

# 68. Model Drift

Monitor whether the relationship between:

```text
Features
```

and:

```text
Outcomes
```

changes over time.

Ride-hailing environments can change because of:

```text
Seasonality
New Roads
New Drivers
New Ride Types
Traffic Patterns
Pricing Changes
Market Expansion
Regulation
```

---

# 69. Data Drift

Monitor changes in feature distributions.

Examples:

```text
Average Pickup Distance
Average Driver Density
Demand Distribution
ETA Distribution
```

Unexpected changes may indicate data drift.

---

# 70. Prediction Drift

Monitor model outputs over time.

Examples:

```text
Score Distribution
Ranking Distribution
Confidence Distribution
```

Sudden changes may indicate:

```text
Feature Problems
Model Problems
Traffic Changes
```

---

# 71. Outcome Monitoring

Model quality must ultimately be evaluated against real outcomes.

Examples:

```text
Pickup ETA
Assignment Success
Driver Acceptance
Cancellation
Driver Utilization
```

---

# 72. AI Is Not Optimized in Isolation

A model with excellent offline metrics may produce poor operational outcomes.

Therefore evaluate:

```text
Model Metrics
+
Dispatch Metrics
+
Business Metrics
+
Safety Metrics
```

---

# 73. Offline Evaluation

Before production:

```text
Historical Data
      ↓
Candidate Generation
      ↓
Model Prediction
      ↓
Offline Evaluation
```

Potential metrics:

```text
Ranking Quality
Top-K Accuracy
NDCG
Expected ETA
Assignment Probability
Calibration
```

The exact metric set depends on the model objective.

---

# 74. Online Evaluation

After controlled deployment:

```text
Real Traffic
      ↓
Monitor Outcomes
```

Compare:

```text
Control
vs
AI Treatment
```

---

# 75. AI Experimentation

AI dispatch changes should be evaluated through controlled experiments where practical.

Experiments must define:

```text
Control
Treatment
Population
Metrics
Duration
Guardrails
Rollback
```

---

# 76. No Uncontrolled Production Experiment

Do not change a dispatch model globally without:

```text
Validation
Monitoring
Rollback
```

---

# 77. Regional Model Selection

Different markets may eventually require different models.

For example:

```text
Region A → Model A
Region B → Model B
```

The architecture should support policy-based model selection.

---

# 78. Model Selection Policy

Model selection may depend on:

```text
Region
Vehicle Category
Ride Type
Dispatch Mode
Experiment Assignment
Model Availability
```

The selected model must remain governed and approved.

---

# 79. Global Default Model

There should be a safe default.

Conceptually:

```text
Configured Model
     ↓
Unavailable?
     ↓
Approved Fallback Model
     ↓
Deterministic Smart Ranking
```

---

# 80. Model Version Pinning

Production dispatch should use an explicitly selected model version.

Avoid:

```text
latest
```

as an implicit production dependency.

Prefer:

```text
dispatch-ranker-v4
```

with explicit promotion.

---

# 81. Model Promotion

A model should move through controlled states:

```text
EXPERIMENTAL
VALIDATED
SHADOW
CANARY
PRODUCTION
DEPRECATED
RETIRED
```

---

# 82. Model Rollback

Rollback should restore a previously approved model without requiring code changes to the dispatch domain.

---

# 83. AI Auditability

The system should be able to answer:

```text
Which model made this prediction?
Which policy was active?
Which feature set was used?
Was fallback used?
What was the outcome?
```

subject to privacy and retention policies.

---

# 84. AI Explainability

For operational purposes, the system should retain a concise explanation context such as:

```text
Strategy
Model Version
Policy Version
Fallback State
Decision Timestamp
```

The system does not need to expose full model internals.

---

# 85. Sensitive Features

AI models must not use sensitive or prohibited information merely because it improves predictive performance.

Feature governance must define:

```text
Allowed
Restricted
Prohibited
```

features.

---

# 86. Location Data

Driver and rider location are operationally important but sensitive.

Their use must comply with:

```text
Data Privacy
Retention
Access Control
Purpose Limitation
```

requirements.

---

# 87. Fairness Constraints

The model must not introduce unfair driver treatment based on inappropriate attributes.

Fairness review should consider:

```text
Assignment Opportunity
Idle Time
Trip Distribution
Earnings
Cancellation
```

where relevant.

---

# 88. AI Safety Boundary

AI must never:

```text
Override Legal Rules
Override Driver Eligibility
Override Vehicle Restrictions
Override Ride State
Bypass Authorization
Directly Modify Production State
```

---

# 89. AI Output as Untrusted Input

Model output should be treated as untrusted application input.

Validate:

```text
Schema
Candidate Identity
Range
Type
Version
Completeness
```

before use.

---

# 90. Model Security

Model-serving infrastructure must protect:

```text
Model Artifacts
Inference APIs
Feature Data
Credentials
Training Data
```

---

# 91. Model Integrity

Production should verify that the deployed model artifact corresponds to the approved version.

Avoid deploying untracked model binaries.

---

# 92. AI Dependency Security

AI services must use controlled dependencies and follow the platform's security requirements.

---

# 93. AI Cost Controls

The architecture should support:

```text
Per-Request Limits
Per-Region Limits
Model Tiering
Traffic Percentage
Fallback
Caching Where Valid
```

---

# 94. AI Cost vs Benefit

AI should remain enabled only when it demonstrates meaningful value.

Measure:

```text
Cost per Inference
Cost per Assignment
Improvement in Pickup ETA
Improvement in Assignment Rate
Reduction in Cancellation
```

---

# 95. AI Model Size

The most accurate model is not automatically the best production model.

Consider:

```text
Accuracy
Latency
Cost
Memory
Throughput
Operational Complexity
```

---

# 96. Model Optimization

Potential optimizations include:

```text
Smaller Model
Quantization
Batching Where Compatible
Candidate Reduction
Feature Caching
Model Warm-Up
Efficient Serialization
```

Only introduce optimizations supported by measured bottlenecks.

---

# 97. Real-Time Feature Caching

Frequently reused features may be cached when their freshness requirements allow it.

Example:

```text
Driver State
Demand Signal
Zone Statistics
```

The cache must never override authoritative state.

---

# 98. AI and Redis

Redis may support real-time feature or state access where appropriate.

The AI architecture must follow:

```text
ADR-0009 — Redis for Real-Time State and Caching
```

and:

```text
ADR-0010 — Driver Location Storage Strategy
```

---

# 99. AI and PostgreSQL

Historical training data and authoritative business data may originate from PostgreSQL.

The AI pipeline must not turn analytical workloads into uncontrolled production database load.

---

# 100. AI and Redpanda

Events may feed:

```text
Training Data Pipelines
Feature Pipelines
Model Monitoring
Outcome Tracking
```

through Redpanda.

The event architecture remains governed by:

```text
ADR-0005
ADR-0006
ADR-0012
ADR-0013
```

---

# 101. Feedback Loop

AI dispatch should learn from outcomes.

Conceptually:

```text
Prediction
   ↓
Dispatch
   ↓
Actual Outcome
   ↓
Feedback Data
   ↓
Evaluation / Training
   ↓
New Model
```

The feedback loop must avoid automatically promoting every learned model into production.

---

# 102. Outcome Examples

Useful outcomes include:

```text
Driver Accepted
Driver Rejected
Offer Expired
Pickup ETA
Ride Cancelled
Ride Completed
Actual Pickup Time
Assignment Success
```

---

# 103. Feedback Data Quality

Training data should be filtered for:

```text
Invalid Events
Duplicate Events
Corrupt Data
Known Operational Incidents
Incomplete Outcomes
```

according to the ML data pipeline rules.

---

# 104. Causal Considerations

Observed outcomes can be influenced by the dispatch strategy itself.

For example:

```text
Driver Acceptance
```

depends partly on:

```text
Which Driver Received the Offer
```

Therefore, model evaluation must avoid naïvely treating historical assignments as unbiased examples.

---

# 105. Exploration vs Exploitation

AI dispatch may eventually need to balance:

```text
Exploration
```

and:

```text
Exploitation
```

However, exploration must be constrained by:

```text
Safety
Fairness
Operational Rules
Business Limits
```

This should be introduced only when justified.

---

# 106. AI and Stand Dispatch

AI should not automatically replace Stand Dispatch.

A stand-based operating area may continue using:

```text
Stand Queue
```

even if AI exists.

AI can be used for:

```text
Demand Forecasting
Queue Planning
Operational Analytics
```

without controlling individual stand assignments.

---

# 107. AI-Assisted Stand Operations

Future AI may assist stand operations by predicting:

```text
Demand
Queue Pressure
Expected Wait
Driver Supply
```

while preserving:

```text
Queue Rules
```

as the actual assignment mechanism.

---

# 108. AI and Smart Dispatch

AI is most naturally integrated with Smart Dispatch:

```text
Smart Candidate Set
        ↓
AI Ranking
        ↓
Policy Validation
        ↓
Assignment
```

This is the primary intended AI dispatch use case.

---

# 109. AI and ETA

ETA is an important dispatch feature.

AI may eventually predict:

```text
Pickup ETA
Acceptance Probability
Assignment Success
```

but route-provider and ETA architecture remain independent.

---

# 110. AI and Demand Prediction

Demand prediction may provide features such as:

```text
Expected Requests
Expected Supply
Zone Pressure
```

for Smart Dispatch.

Demand prediction must remain separate from the assignment transaction.

---

# 111. AI and Driver Supply

Supply prediction may estimate:

```text
Available Drivers
Expected Driver Arrivals
Expected Driver Departures
```

These are optimization signals, not authoritative driver state.

---

# 112. AI and Matching

AI ranking should remain compatible with the matching architecture.

The flow is:

```text
Candidate Discovery
   ↓
Eligibility
   ↓
AI Ranking
   ↓
Assignment
```

---

# 113. AI Does Not Own Matching State

The AI model must not maintain authoritative:

```text
Driver Availability
Ride State
Assignment State
```

---

# 114. AI Observability

Monitor:

```text
Inference Latency
Inference Errors
Fallback Rate
Model Version
Feature Errors
Prediction Distribution
Data Drift
Model Drift
Business Outcomes
```

---

# 115. AI Alerts

Potential alerts:

```text
High Inference Failure Rate
High Fallback Rate
High Latency
Model Version Error
Feature Availability Drop
Prediction Distribution Shift
Business KPI Regression
```

---

# 116. Fallback Rate

A high AI fallback rate may indicate:

```text
Model Service Instability
Feature Problems
Capacity Issues
Timeout Configuration
```

Fallback is safe, but excessive fallback means the AI path is not providing its intended value.

---

# 117. AI Incident Response

When AI behaves unexpectedly:

```text
Detect
 ↓
Evaluate Impact
 ↓
Disable AI / Roll Back
 ↓
Use Deterministic Smart Dispatch
 ↓
Investigate
 ↓
Fix
 ↓
Shadow Validate
 ↓
Controlled Re-Enable
```

---

# 118. Kill Switch

The architecture should support an operational AI kill switch.

Conceptually:

```text
AI_ENABLED = false
```

or equivalent policy configuration.

Disabling AI should not require redeploying the core dispatch service if the configuration architecture supports safe runtime changes.

---

# 119. Kill Switch Safety

The kill switch should:

```text
Be Authorized
Be Audited
Be Observable
Have Safe Defaults
```

---

# 120. Model-Specific Disable

The platform should be able to disable one model version without disabling all Smart Dispatch.

Example:

```text
dispatch-ranker-v4
→ Disabled

dispatch-ranker-v3
→ Approved
```

---

# 121. Regional AI Disable

AI may be disabled for one region while remaining enabled elsewhere.

Example:

```text
Region A → AI Enabled
Region B → AI Disabled
```

The deterministic Smart strategy remains available.

---

# 122. Deployment Independence

Model deployment should be independent from core dispatch application deployment where practical.

This allows:

```text
Model Update
```

without requiring:

```text
Core Dispatch Redeployment
```

---

# 123. Contract Stability

The AI inference contract should remain stable even when the internal model changes.

For example:

```text
RankingRequest
RankingResponse
```

should remain independent of model framework details.

---

# 124. AI API Contract

A conceptual request:

```text
{
  "request_id": "...",
  "ride_id": "...",
  "model_version": "...",
  "candidates": [...]
}
```

A conceptual response:

```text
{
  "model_version": "...",
  "ranked_candidates": [...],
  "inference_time_ms": 0
}
```

The actual schema belongs to the AI service implementation.

---

# 125. AI Contract Validation

The dispatch service must validate:

```text
Request Serialization
Response Serialization
Candidate IDs
Model Version
Score Values
```

---

# 126. Backward Compatibility

Model-serving API changes should remain backward-compatible where possible.

If a breaking change is required:

```text
Version the inference contract
```

rather than silently changing behaviour.

---

# 127. AI Service Authentication

The dispatch service must authenticate to the AI inference service.

The AI service must not assume that any internal caller is trusted.

---

# 128. AI Authorization

Only approved services should be able to invoke production dispatch models.

---

# 129. AI Request Rate Limits

The AI service should enforce appropriate limits to prevent accidental overload.

---

# 130. AI Data Residency

If model inference is performed outside the primary deployment environment, data transfer requirements must be evaluated before production use.

---

# 131. External AI Providers

External AI providers should not receive sensitive ride or driver data unless explicitly approved by the applicable data governance and security policies.

---

# 132. Model Governance

Every production model should have:

```text
Owner
Purpose
Training Data Reference
Feature Set
Evaluation Metrics
Approval
Version
Deployment Status
Rollback Version
```

---

# 133. Model Retirement

Models should be retired when:

```text
Performance Degrades
Better Model Exists
Feature Contract Changes
Business Objective Changes
Operational Cost Is Too High
```

---

# 134. Model Reproducibility

Where practical, preserve:

```text
Model Artifact
Training Configuration
Feature Version
Evaluation Dataset Reference
Code Version
Model Version
```

to support reproducibility.

---

# 135. Model Registry Metadata

At minimum:

```text
model_name
model_version
status
created_at
approved_at
feature_version
code_version
evaluation_summary
```

---

# 136. AI Deployment Environments

Use separate model environments:

```text
Development
Staging
Production
```

Production models should not be changed directly from local development.

---

# 137. Model Promotion

Promotion should follow:

```text
Development
    ↓
Evaluation
    ↓
Staging
    ↓
Shadow
    ↓
Canary
    ↓
Production
```

---

# 138. AI Development Workflow

A typical workflow is:

```text
Problem Definition
      ↓
Data Preparation
      ↓
Feature Engineering
      ↓
Training
      ↓
Offline Evaluation
      ↓
Model Registration
      ↓
Serving Validation
      ↓
Shadow Deployment
      ↓
Canary
      ↓
Production
```

---

# 139. AI Development Checklist

Before enabling an AI dispatch model:

```text
[ ] Objective defined
[ ] Features defined
[ ] Data quality validated
[ ] Offline metrics validated
[ ] Latency measured
[ ] Cost measured
[ ] Safety constraints verified
[ ] Fairness evaluated
[ ] Model registered
[ ] Version pinned
[ ] Shadow-tested
[ ] Fallback tested
[ ] Rollback tested
[ ] Monitoring configured
[ ] Ownership assigned
```

---

# 140. Production Readiness

The AI path is production-ready only when:

```text
Model Is Approved
+
Inference Is Reliable
+
Fallback Works
+
Observability Works
+
Rollback Works
+
Hard Constraints Are Enforced
```

---

# 141. Consequences

## 141.1 Positive Consequences

The decision provides:

```text
Better Candidate Ranking Potential
Controlled AI Adoption
Deterministic Fallback
Independent Model Deployment
Model Versioning
Experimentation
Operational Kill Switch
Regional AI Control
```

---

## 141.2 Negative Consequences

The architecture introduces:

```text
Model Serving Infrastructure
Feature Engineering
Model Governance
Monitoring
Training Pipelines
AI Operational Cost
Additional Failure Modes
```

These trade-offs are accepted because AI remains optional and bounded.

---

# 142. Risks

## Risk 1 — AI Becomes a Single Point of Failure

### Mitigation

```text
Deterministic Smart Fallback
Timeout
Circuit Breaker
Kill Switch
```

---

## Risk 2 — Model Optimizes the Wrong Objective

### Mitigation

Use:

```text
Explicit Objective
Business Metrics
Guardrails
Offline Evaluation
Online Experimentation
```

---

## Risk 3 — Model Learns Unfair Behaviour

### Mitigation

Use:

```text
Feature Governance
Fairness Metrics
Policy Constraints
Human Review
Controlled Rollout
```

---

## Risk 4 — Model Drift

### Mitigation

Monitor:

```text
Data Drift
Prediction Drift
Outcome Metrics
```

and retrain or roll back when necessary.

---

## Risk 5 — Inference Cost Becomes Excessive

### Mitigation

Use:

```text
Candidate Reduction
Model Optimization
Traffic Controls
Caching Where Safe
Fallback
```

---

## Risk 6 — Stale Features Produce Bad Decisions

### Mitigation

Use:

```text
Feature Freshness Checks
Feature Timestamps
Fallback
```

---

## Risk 7 — AI Output Causes Invalid Assignment

### Mitigation

Treat model output as untrusted and perform final eligibility and state validation.

---

# 143. Validation

This ADR should be validated through:

```text
Offline Model Evaluation
Shadow Testing
Inference Integration Tests
Failure Tests
Timeout Tests
Fallback Tests
Concurrency Tests
Load Tests
Fairness Tests
Drift Monitoring
Canary Deployment
Rollback Tests
```

---

# 144. Review Triggers

Revisit this ADR when:

```text
AI Becomes Mandatory
A New Dispatch Model Is Introduced
A New Model Serving Platform Is Introduced
Model Latency Changes Significantly
Dispatch Scale Changes Significantly
AI Cost Becomes Material
New Data Governance Requirements Appear
A New AI Provider Is Introduced
Reinforcement Learning Is Considered
```

---

# 145. Related Documentation

This ADR should be read together with:

```text
03-components/
04-development/
05-ai/
```

Especially:

```text
AI Strategy and Vision
AI Architecture
AI Components and Services
Smart Dispatch AI
ETA and Prediction System
Driver Demand and Supply Prediction
AI Matching and Ranking
Feature Engineering
Machine Learning Data Pipeline
Model Training and Evaluation
Model Serving and Inference
Model Versioning and Registry
Online and Offline Features
AI Feedback and Learning Loop
AI Safety and Guardrails
AI Monitoring and Model Observability
AI Performance and Cost Optimization
AI Experimentation and A/B Testing
AI Data Privacy and Governance
AI Failure and Fallback Strategy
```

---

# 146. Related ADRs

This decision is directly related to:

```text
ADR-0002 — Architecture Style
ADR-0003 — Microservice Boundaries
ADR-0004 — Domain-Driven Design
ADR-0005 — Event-Driven Architecture
ADR-0006 — Kafka / Redpanda for Event Streaming
ADR-0008 — PostGIS for Geospatial Operations
ADR-0009 — Redis for Real-Time State and Caching
ADR-0010 — Driver Location Storage Strategy
ADR-0013 — Dead Letter Queue Strategy
ADR-0014 — API and Service Communication
ADR-0015 — Smart Dispatch and Stand Dispatch
ADR-0017 — ETA and Route Provider Strategy
ADR-0018 — Regional and Legal Ride Validation
ADR-0019 — Data Consistency and Transaction Boundaries
ADR-0020 — Idempotency Strategy
ADR-0021 — Failure and Degradation Strategy
ADR-0022 — Observability Strategy
ADR-0023 — Security and Secret Management
ADR-0024 — Configuration and Environment Strategy
ADR-0025 — Testing and Integration Strategy
ADR-0026 — Model and AI Governance
ADR-0028 — Cost Optimization Strategy
```

---

# 147. Decision Summary

RideForge will use AI as an optional ranking layer inside Smart Dispatch:

```text
                       Ride Request
                            │
                            ▼
                    Ride / Policy Validation
                            │
                            ▼
                    Regional / Legal Rules
                            │
                            ▼
                    Candidate Discovery
                            │
                            ▼
                  Hard Eligibility Filtering
                            │
                            ▼
                    Candidate Feature Set
                            │
                            ▼
                    AI Ranking Service
                       /          \
                 Success          Failure
                    │               │
                    ▼               ▼
              AI Ranking      Deterministic
                              Smart Ranking
                    │               │
                    └───────┬───────┘
                            ▼
                    Final Revalidation
                            │
                            ▼
                       Reservation
                            │
                            ▼
                        Assignment
                            │
                            ▼
                     DriverAssigned
```

The architectural boundaries are:

```text
Domain Rules
→ Deterministic

Candidate Eligibility
→ Deterministic

AI Ranking
→ Advisory / Optimization

Final Assignment
→ Authoritative Application Logic

Event Publication
→ Transactional Outbox

Failure Handling
→ Deterministic Fallback
```

---

# 148. Final Principle

> **AI should improve dispatch decisions without becoming responsible for dispatch correctness.**

The responsibility hierarchy is:

```text
Business / Legal Rules
        ↓
Hard Eligibility
        ↓
Dispatch Policy
        ↓
AI / Deterministic Ranking
        ↓
Final State Validation
        ↓
Atomic Assignment
        ↓
Event Publication
```

AI is therefore an optimization component rather than an authority.

The system must remain capable of:

```text
Operating Without AI
Rolling Back AI
Disabling AI
Changing Models
Comparing Models
Auditing Decisions
Recovering From Model Failure
```

without redesigning the core dispatch system.

---

# 149. Status

```text
Decision: ACCEPTED

AI Role:
Optional Dispatch Intelligence

Primary Use:
Smart Dispatch Candidate Ranking

Hard Constraints:
Always Deterministic

AI Dependency:
Non-Mandatory

Failure Strategy:
Deterministic Smart Dispatch Fallback

Model Versioning:
Required

Feature Governance:
Required

Observability:
Required

Production Rollout:
Shadow → Canary → Controlled Production

Rollback:
Required

Kill Switch:
Required

Primary Goal:
Improve Dispatch Quality Without Sacrificing Safety, Reliability, or Control
```

This decision establishes the AI-assisted dispatch boundary for RideForge and provides a controlled path from deterministic Smart Dispatch to progressively more intelligent dispatch while keeping the core ride-assignment system reliable and operationally safe.

---

# 24. Clarification: AI-Assisted Dispatch and the Two Primary Strategies

The authoritative dispatch model consists of two primary dispatch strategies:

```text
1. Smart Stand Dispatch
2. Smart Dispatch
```

**AI-Assisted Dispatch is not a third primary dispatch strategy.**

It is an optimization capability that can operate within either primary strategy.

```text
Resolved Dispatch Strategy
        ↓
Strategy-Specific Dispatch
        ↓
AI-Assisted Optimization, where enabled
        ↓
Final Candidate Selection
```

The configured primary strategy remains authoritative even when AI assistance is enabled or unavailable.

---

## 24.1 Hierarchical Dispatch Strategy Resolution

The effective dispatch strategy must be resolved before strategy execution.

Configuration may exist at different levels, including:

```text
State
District
City / Town
Rural Area
Auto Stand
Specific Ride Level
Other configured intermediate levels
```

Not every level requires explicit configuration.

The system must start at the most specific applicable level and move upward until it finds an explicit dispatch strategy.

```text
Most Specific Applicable Level
        ↓
Explicit Strategy?
   ├── YES → Effective Strategy
   └── NO
        ↓
Parent Level
        ↓
Continue Upward
        ↓
System Default if No Explicit Strategy Exists
```

The canonical rule is:

> **Specific configuration overrides inherited configuration.**

The hierarchy must not be implemented as a permanently hard-coded list of geographic levels.

---

## 24.2 Smart Stand Dispatch with AI Assistance

Smart Stand Dispatch is **stand-preferred, not stand-exclusive**.

When the rider is within the configured radius of an auto stand:

```text
Ride Pickup
    ↓
Preferred Auto Stand
    ↓
Eligible Drivers at Preferred Stand
    ↓
Stand Queue / Ordering Rules
    ↓
Suitable Driver?
```

If suitable stand supply is unavailable, the dispatch process may consider broader eligible candidates, including:

```text
Drivers outside the preferred stand
Drivers at nearby stands
Drivers from nearby locations
```

AI may assist with ranking or prediction where the applicable business rules permit it.

However, AI must not turn Smart Stand Dispatch into a generic stand-only scoring system.

If the stand rule is first-eligible-in-queue, AI ranking must not silently override that rule.

---

## 24.3 Smart Stand Dispatch Outside Stand Radius

If the rider is outside the radius of all configured auto stands, Smart Stand Dispatch must not create a stand-only candidate pool.

Eligible nearby drivers may be considered regardless of stand membership.

```text
No Applicable Stand
        ↓
Eligible Nearby Drivers
        ↓
Strategy-Specific Processing
        ↓
AI Assistance Where Enabled
```

---

## 24.4 Smart Dispatch with AI Assistance

Smart Dispatch is stand-agnostic.

The candidate pool may include any eligible nearby driver regardless of whether the driver is:

```text
At an auto stand
Outside an auto stand
Associated with another nearby location
```

AI may assist with:

```text
Candidate Ranking
ETA Prediction
Acceptance Prediction
Demand / Supply Signals
Other Approved Prediction Signals
```

Stand membership must not become an implicit ranking advantage merely because AI assistance is enabled.

---

## 24.5 Cross-Location Dispatch

Dispatch candidate discovery may expand to nearby locations when the originating location cannot provide a suitable driver.

Example:

```text
Location A
Smart Dispatch
    ↓
No suitable local candidate
    ↓
Location B
Smart Stand Dispatch
    ↓
Eligible candidates from B
```

Drivers from Location B may be considered even though B uses a different primary dispatch strategy.

The source context must be preserved:

```text
Candidate Location
Candidate Location Strategy
Stand Membership
Relevant Stand
Queue Position
Discovery Source
Expansion Level
```

The candidate is ultimately evaluated according to the applicable eligibility and dispatch/ranking rules.

A different source-location strategy is not, by itself, a reason to reject the candidate.

---

## 24.6 AI Is Not a Candidate Boundary

AI-assisted dispatch must not redefine geographic or operational candidate boundaries.

These concepts remain separate:

```text
Dispatch Strategy
Candidate Discovery Scope
Eligibility
Strategy-Specific Prioritization
AI Ranking
```

Therefore:

```text
AI-Assisted Smart Stand Dispatch
        ≠
Stand-Only Candidate Pool
```

and:

```text
AI-Assisted Smart Dispatch
        ≠
AI May Ignore Hard Constraints
```

---

## 24.7 Hard Constraints Always Take Precedence

AI assistance must operate below authoritative constraints.

AI cannot override:

```text
Legal / Regional Restrictions
Driver Eligibility
Driver Availability
Vehicle / Service Compatibility
Safety Constraints
Ride Constraints
Location Freshness Requirements
Other Hard Business Rules
```

The AI output is advisory/optimization input within the allowed candidate space.

---

## 24.8 AI Failure and Deterministic Fallback

If AI assistance is unavailable, the primary dispatch strategy remains unchanged.

For Smart Stand Dispatch:

```text
Smart Stand Dispatch
        ↓
AI unavailable
        ↓
Deterministic Smart Stand Dispatch
```

For Smart Dispatch:

```text
Smart Dispatch
        ↓
AI unavailable
        ↓
Deterministic Smart Dispatch
```

AI failure must not silently cause:

```text
Smart Stand Dispatch
        ↓
Smart Dispatch
```

unless an explicit business/configuration rule defines that transition.

Similarly, broader candidate discovery must not be interpreted as a strategy switch.

---

## 24.9 Strategy Context for AI

AI models and ranking components should receive sufficient context to distinguish the dispatch situation.

Relevant context may include:

```text
Effective Dispatch Strategy
Configuration Level
Applicable Stand
Stand Membership
Stand Queue Position
Candidate Location
Candidate Location Strategy
Discovery Source
Expansion Level
Distance
ETA
Driver Availability
Other Approved Features
```

The model must not infer strategy solely from stand membership or geographic location.

The authoritative effective strategy comes from configuration resolution.

---

## 24.10 Strategy and AI Observability

Dispatch decisions should record AI assistance separately from the primary strategy.

For example:

```text
strategy = SMART
ai_assisted = true
model_version = dispatch-ranker-vX
```

or:

```text
strategy = SMART_STAND
ai_assisted = false
```

This preserves the distinction between:

```text
Primary Dispatch Strategy
```

and:

```text
AI Assistance
```

---

## 24.11 AI-Assisted Dispatch Guardrails

Implementations must not:

```text
Treat AI-Assisted Dispatch as a third primary dispatch strategy.

Treat Smart Stand Dispatch as stand-only.

Reject non-stand candidates solely because the rider is inside a stand radius.

Reject nearby-location candidates solely because their source location uses another dispatch strategy.

Resolve dispatch configuration inside the AI model.

Hard-code the configuration hierarchy.

Allow AI to override hard eligibility, legal, safety, or operational constraints.

Replace a configured stand queue rule with an arbitrary AI score.

Switch Smart Stand Dispatch to Smart Dispatch merely because AI is unavailable.

Treat candidate expansion as strategy switching.
```

The implementation must preserve the resolved primary dispatch strategy and use AI only as an optimization capability within the permitted candidate and business-rule boundaries.

