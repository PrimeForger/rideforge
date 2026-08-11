# ADR-0026: Model and AI Governance

> **Status:** Accepted  
> **Date:** 2026-08-09  
> **Decision Type:** AI / Machine Learning / Governance / Safety  
> **Scope:** AI models, machine learning models, model lifecycle, model approval, model versioning, training data, evaluation, deployment, inference, monitoring, rollback, AI-assisted dispatch, ETA prediction, experimentation, explainability, safety, privacy, and operational governance  
> **Owner:** RideForge Architecture / AI Engineering  
> **Supersedes:** None  
> **Superseded By:** None

---

# 1. Context

RideForge includes AI and machine-learning capabilities intended to improve:

```text
Smart Dispatch
Driver Matching
Driver Ranking
ETA Prediction
Demand Prediction
Supply Prediction
Operational Optimization
```

AI is therefore part of the operational platform rather than an isolated experimental feature.

AI-generated decisions may influence:

```text
Which Driver Is Considered
Which Driver Is Ranked Higher
Which Dispatch Strategy Is Used
Estimated Arrival Time
Demand Forecast
Supply Forecast
Operational Recommendations
```

Incorrect AI behaviour can therefore affect:

```text
Customer Experience
Driver Experience
Operational Efficiency
Platform Reliability
Fairness
Safety
Legal / Regional Compliance
Cost
Revenue
```

AI must consequently operate inside explicit deterministic platform constraints.

---

# 2. Problem

Without governance, AI systems can gradually become difficult to control.

Potential problems include:

```text
Unapproved Models
Untracked Model Versions
Unvalidated Training Data
Hidden Model Changes
Poor Reproducibility
Data Leakage
Model Drift
Feature Drift
Unsafe Predictions
Unbounded AI Authority
Vendor Lock-In
Unexpected Cost
Poor Rollback
Difficult Incident Investigation
```

RideForge requires a governance model defining:

```text
Which models are allowed
How models are approved
How models are versioned
How models are evaluated
How models are deployed
How models are monitored
How models are rolled back
What AI is allowed to control
What AI is never allowed to override
```

---

# 3. Decision

RideForge will adopt a:

```text
Controlled
Versioned
Evaluated
Observable
Reversible
Constraint-Bounded
```

AI governance model.

The core principle is:

> **AI may optimize decisions inside explicitly defined business, legal, security, and safety boundaries, but AI must not become the authority that defines or bypasses those boundaries.**

---

# 4. AI Authority Model

AI operates inside a deterministic control boundary.

```text
                    Request
                       │
                       ▼
              Deterministic Rules
                       │
                       ▼
              Eligible Candidates
                       │
                       ▼
                 AI / ML Layer
                       │
                       ▼
              Prediction / Ranking
                       │
                       ▼
             Business Validation
                       │
                       ▼
                Final Decision
```

AI is therefore:

```text
Optimizer
Predictor
Ranker
Estimator
```

and not:

```text
Legal Authority
Security Authority
Identity Authority
Final Safety Authority
```

---

# 5. Hard Constraints

Hard constraints must be enforced outside the model where practical.

Examples:

```text
Driver Eligibility
Ride Eligibility
Regional Restrictions
Legal Restrictions
Account Status
Vehicle Requirements
Safety Requirements
Resource Availability
Authorization
```

---

# 6. AI Cannot Override Legal Rules

AI must never override:

```text
Regional Ride Validation
Legal Operating Restrictions
Mandatory Regulatory Rules
```

These remain deterministic domain rules.

This is aligned with:

```text
ADR-0018 — Regional and Legal Ride Validation
```

---

# 7. AI Cannot Override Authorization

A model must never cause an unauthorized action merely because its score or recommendation is high.

---

# 8. AI Cannot Override Safety

Safety constraints remain outside the model.

Example:

```text
AI Score = 0.99
```

does not make an otherwise ineligible driver eligible.

---

# 9. AI Model Categories

RideForge AI models are conceptually classified as:

```text
Prediction Models
Ranking Models
Classification Models
Forecasting Models
Optimization Models
Recommendation Models
Generative / LLM Components
```

---

# 10. Initial High-Value Models

The primary AI/ML areas are:

```text
Smart Dispatch
ETA Prediction
Driver Demand Prediction
Supply Prediction
Driver Matching / Ranking
```

---

# 11. Model Registry

Every production model must have a traceable identity.

A model record should include:

```text
Model Name
Model Version
Model Type
Training Dataset Version
Feature Version
Code Version
Evaluation Version
Deployment Status
Approval Status
Owner
Created At
Approved At
```

---

# 12. Model Versioning

Models must be immutable once deployed.

If a model changes materially:

```text
New Model Version
```

must be created.

Do not silently overwrite a production model artifact.

---

# 13. Model Naming

Model identifiers should be predictable.

Example:

```text
dispatch-ranker-v1
eta-predictor-v3
demand-forecast-v2
```

---

# 14. Model Artifact Integrity

Production model artifacts should be traceable to:

```text
Source Code
Training Data Version
Feature Definition
Training Configuration
Evaluation Results
```

---

# 15. Reproducibility

A production model should be reproducible or reconstructable from its recorded inputs and training metadata where practical.

---

# 16. Training Data Governance

Training data must have a known origin.

Examples:

```text
Ride History
Driver History
Location History
Traffic / Routing Data
Operational Events
Historical ETA
Dispatch Outcomes
```

---

# 17. Training Data Ownership

Each important training dataset should have an owner.

---

# 18. Training Dataset Versioning

Material changes to training datasets should result in a new dataset version.

---

# 19. Data Leakage

Training pipelines must prevent future information from leaking into historical training examples.

Example:

```text
Future Ride Outcome
```

must not be used as a feature when predicting the state before that outcome occurred.

---

# 20. Temporal Validation

Time-dependent ride-hailing models should use temporal validation where appropriate.

Example:

```text
Historical Period
→ Training

Later Period
→ Validation

Future Period
→ Evaluation
```

---

# 21. Geographic Validation

Models whose behaviour depends on geography should be evaluated across relevant operating regions.

---

# 22. Regional Generalization

A model performing well in one region must not automatically be assumed to perform well in another.

---

# 23. Cold-Start Behaviour

Models must define behaviour when:

```text
New Region
New Driver
New Vehicle
New Route
Sparse Historical Data
```

is encountered.

---

# 24. Missing Features

Models must define behaviour when required features are unavailable.

Possible behaviour:

```text
Fallback Prediction
Default Score
Alternative Model
Deterministic Algorithm
Reject AI Path
```

The choice must be explicit.

---

# 25. Feature Governance

Production model features must be:

```text
Defined
Versioned
Validated
Observable
Owned
```

---

# 26. Feature Definitions

A feature should have a documented meaning.

Example:

```text
driver_recent_acceptance_rate
```

must have an explicit definition covering:

```text
Time Window
Included Events
Calculation
Missing Value Behaviour
```

---

# 27. Feature Versioning

Material changes to feature semantics require a new feature version.

---

# 28. Online Features

Online features must have defined:

```text
Source
Freshness
TTL
Fallback
Validation
```

---

# 29. Offline Features

Offline training features must be generated consistently with production semantics where the same feature is used online.

---

# 30. Training / Serving Skew

RideForge must actively avoid:

```text
Training Feature Definition
≠
Production Feature Definition
```

when they are intended to represent the same feature.

---

# 31. Model Evaluation

A model must meet predefined evaluation criteria before production deployment.

Evaluation should include both:

```text
Model Metrics
+
Business Metrics
```

---

# 32. Model Metrics

Depending on model type, metrics may include:

```text
MAE
RMSE
Precision
Recall
F1
AUC
Calibration
Ranking Metrics
Forecast Error
Latency
```

---

# 33. Business Metrics

Business evaluation may include:

```text
Driver Acceptance Rate
Customer Wait Time
Pickup ETA Accuracy
Dispatch Success Rate
Cancellation Rate
Deadhead Distance
Driver Utilization
Marketplace Balance
```

---

# 34. Offline Evaluation

Offline evaluation is required before production deployment where sufficient historical data exists.

---

# 35. Online Evaluation

Models should be evaluated in production-like conditions through:

```text
Shadow Mode
Canary
A/B Test
Limited Rollout
```

where appropriate.

---

# 36. Shadow Mode

A new model may receive production inputs and produce predictions without controlling actual decisions.

Example:

```text
Production Model
→ Controls Dispatch

Candidate Model
→ Predicts in Shadow

Compare Results
```

---

# 37. Canary Deployment

A new model may initially receive a small percentage of traffic.

Example:

```text
Existing Model → 95%
Candidate Model → 5%
```

The exact allocation depends on risk.

---

# 38. A/B Testing

A/B testing may be used where the model change is appropriate for controlled experimentation.

Experimentation follows:

```text
05-ai documentation
```

and the relevant experimentation strategy.

---

# 39. Model Approval

A production model requires explicit approval.

Approval should consider:

```text
Offline Performance
Business Impact
Safety
Legal Constraints
Cost
Latency
Operational Readiness
Rollback
```

---

# 40. Approval Record

The approval record should identify:

```text
Model Version
Evaluator
Evaluation Date
Dataset Version
Decision
Known Limitations
Rollback Plan
```

---

# 41. Model Lifecycle

The model lifecycle is:

```text
Idea
 ↓
Data
 ↓
Feature Engineering
 ↓
Training
 ↓
Evaluation
 ↓
Approval
 ↓
Shadow
 ↓
Canary
 ↓
Production
 ↓
Monitoring
 ↓
Retraining / Replacement
 ↓
Retirement
```

---

# 42. Model Development Environment

Model development should remain isolated from production systems.

---

# 43. Production Data Access

Access to production data for model development must be controlled.

Sensitive data should be:

```text
Minimized
Restricted
Sanitized
Anonymized / Pseudonymized Where Appropriate
```

---

# 44. Data Privacy

AI training and inference must follow the platform's privacy and data governance requirements.

---

# 45. Location Data

Driver and rider location data can be highly sensitive.

AI systems must only use the location data required for the intended model purpose.

---

# 46. Data Retention

Training data should follow defined retention policies.

Do not retain data indefinitely merely because it might be useful for future modeling.

---

# 47. External AI Providers

If an external AI provider is used:

```text
Data Sent
Retention
Provider Access
Region
Security
Cost
```

must be evaluated.

---

# 48. AI Provider Credentials

AI provider credentials follow:

```text
ADR-0023 — Security and Secret Management
```

They must not be exposed to client applications.

---

# 49. Generative AI

Generative AI must not directly perform privileged operational actions without deterministic validation.

---

# 50. LLM Output Validation

LLM output must be treated as untrusted output.

Validate:

```text
Schema
Allowed Values
Business Rules
Security Constraints
```

before using the output.

---

# 51. Prompt Governance

Production prompts should be:

```text
Versioned
Reviewed
Tested
Traceable
```

where prompts materially affect system behaviour.

---

# 52. Prompt Changes

Material prompt changes should be treated similarly to model behaviour changes.

---

# 53. Model Serving

Production inference should expose a controlled interface.

Example:

```text
Application
    ↓
Inference Service
    ↓
Model
    ↓
Validated Prediction
```

---

# 54. Model Service Isolation

Model-serving systems should not automatically receive unrestricted access to:

```text
Database
Payment Systems
Administrative APIs
```

---

# 55. Model Latency

AI inference must have an explicit latency budget.

This is particularly important for:

```text
Real-Time Dispatch
ETA
Matching
```

---

# 56. Model Timeout

Every production inference request must have a bounded timeout.

---

# 57. Model Fallback

If AI fails:

```text
Timeout
Error
Unavailable
Invalid Output
```

the platform must use an approved fallback.

---

# 58. Dispatch Fallback

For smart dispatch, the fallback may be:

```text
Deterministic Ranking
Stand Dispatch
Alternative Dispatch Strategy
```

depending on operating context.

---

# 59. ETA Fallback

ETA may fall back to:

```text
Routing Provider
Deterministic Estimate
Previous Valid Estimate
```

depending on the ETA architecture.

---

# 60. Demand Forecast Fallback

Demand prediction failure may fall back to:

```text
Historical Baseline
Moving Average
Rule-Based Forecast
```

where appropriate.

---

# 61. Model Failure Must Not Become System Failure

AI is an optimization capability.

The core ride lifecycle must remain operational where an approved deterministic fallback exists.

---

# 62. Model Output Validation

Every production model should validate output.

Examples:

```text
Score Range
Probability Range
Prediction Bounds
Required Fields
Finite Numbers
```

---

# 63. Invalid Model Output

If a model produces:

```text
NaN
Infinity
Missing Output
Invalid Class
Out-of-Range Value
```

the prediction must be rejected.

---

# 64. Model Calibration

Probability-producing models should be evaluated for calibration when decisions depend on probability interpretation.

---

# 65. Ranking Model Safety

Ranking models must not rank candidates that have already failed hard eligibility checks.

---

# 66. Candidate Generation Boundary

The recommended architecture is:

```text
Hard Eligibility
       ↓
Candidate Generation
       ↓
Feature Generation
       ↓
Model Ranking
       ↓
Business Validation
       ↓
Assignment
```

---

# 67. Smart Dispatch Governance

Smart dispatch models must be evaluated against:

```text
Pickup ETA
Acceptance Rate
Driver Utilization
Customer Wait
Cancellation
Fairness
Operational Stability
```

---

# 68. Stand Dispatch Governance

Stand dispatch remains deterministic and must remain independently operable.

---

# 69. Hybrid Dispatch Governance

RideForge may select:

```text
Smart Dispatch
Stand Dispatch
```

according to approved regional or operational configuration.

AI must not silently change the dispatch mode.

---

# 70. Model Explainability

Not every model requires human-readable explanations, but production decisions must have sufficient telemetry to understand:

```text
Model Version
Input Context
Output
Fallback
Decision Path
```

subject to privacy constraints.

---

# 71. Decision Trace

Important AI-assisted decisions should be traceable.

Example:

```text
Ride ID
Model Version
Candidate Set
Feature Version
Model Scores
Final Selection
Fallback State
```

Sensitive raw features should not be logged unnecessarily.

---

# 72. Model Observability

Monitor:

```text
Inference Latency
Error Rate
Timeout Rate
Prediction Distribution
Feature Availability
Feature Drift
Model Drift
Fallback Rate
Cost
```

---

# 73. Model Drift

Model performance can degrade as real-world conditions change.

Examples:

```text
Traffic Patterns
Driver Behaviour
Demand Patterns
New Regions
Seasonality
Pricing Changes
Operational Policy Changes
```

---

# 74. Feature Drift

Monitor important feature distributions.

Example:

```text
Historical Distribution
vs
Current Distribution
```

---

# 75. Prediction Drift

Monitor model output distributions.

Unexpected changes may indicate:

```text
Data Problem
Feature Problem
Model Problem
Business Change
```

---

# 76. Performance Drift

Where ground truth becomes available, compare predictions with actual outcomes.

---

# 77. Retraining

Retraining should occur according to:

```text
Performance
Data Freshness
Drift
Business Change
Scheduled Cadence
```

rather than automatically retraining without validation.

---

# 78. Retraining Governance

Every retrained model is a new model version and must pass evaluation before production deployment.

---

# 79. Model Registry Status

Model versions should have lifecycle states such as:

```text
DEVELOPMENT
EVALUATION
APPROVED
SHADOW
CANARY
PRODUCTION
DEPRECATED
RETIRED
REJECTED
```

---

# 80. Production Model Immutability

A model marked:

```text
PRODUCTION
```

must not be modified in place.

---

# 81. Rollback

Every production model deployment must have a rollback path.

---

# 82. Rollback Target

Rollback should identify a previously approved model version.

Example:

```text
eta-v4
↓
rollback
↓
eta-v3
```

---

# 83. Rollback Trigger

Rollback may be triggered by:

```text
Error Rate
Latency
Business Metric Regression
Safety Issue
Model Drift
Provider Failure
Operational Instability
```

---

# 84. Automatic Rollback

Automatic rollback may be introduced for clearly measurable failures.

It should be conservative for business-critical models.

---

# 85. Model Retirement

A model should be retired when:

```text
Replaced
No Longer Useful
Unsupported
Unsafe
Too Expensive
Too Inaccurate
```

---

# 86. Retired Model Artifacts

Historical model metadata should remain available for audit and incident investigation according to retention requirements.

---

# 87. Model Registry Security

Only authorized identities should be able to:

```text
Register
Approve
Deploy
Retire
Delete
```

models.

---

# 88. Model Deployment Authorization

Training access does not automatically imply production deployment permission.

---

# 89. Separation of Duties

Where practical:

```text
Model Developer
≠
Production Approver
```

for high-impact models.

---

# 90. Production Access

Production model deployment should use controlled CI/CD or deployment tooling.

Avoid manual replacement of model artifacts on production hosts.

---

# 91. Model Artifact Storage

Model artifacts should be stored in controlled storage with:

```text
Access Control
Versioning
Integrity
Retention
```

---

# 92. Model Artifact Encryption

Sensitive model artifacts should use encryption at rest where required.

---

# 93. Model Metadata

Model metadata should capture:

```text
Training Code Version
Data Version
Feature Version
Model Version
Evaluation Metrics
Deployment History
```

---

# 94. Model Cost Governance

AI inference has direct operational cost.

Monitor:

```text
Requests
Tokens Where Applicable
Compute Time
GPU / CPU Usage
Provider Charges
```

---

# 95. Cost Budgets

High-cost models should have defined operational limits.

---

# 96. Cost-Aware Model Selection

A more accurate model is not automatically better if:

```text
Latency
Cost
Reliability
```

make it unsuitable for the production workflow.

---

# 97. Real-Time Dispatch Constraint

Dispatch inference must satisfy the latency budget of the ride lifecycle.

AI must not make dispatch unusably slow.

---

# 98. Batch Prediction

Batch inference should be preferred where real-time prediction is not necessary.

---

# 99. Online Inference

Online inference is appropriate when decisions require current state.

Examples:

```text
Dispatch Ranking
Real-Time ETA
```

---

# 100. Offline Inference

Offline inference may be used for:

```text
Demand Forecasting
Model Training
Analytics
Long-Horizon Predictions
```

---

# 101. Model Evaluation Dataset

Evaluation datasets should be:

```text
Versioned
Representative
Protected
Reproducible
```

---

# 102. Evaluation Leakage

Evaluation datasets must not be used to tune models repeatedly without accounting for overfitting to the evaluation set.

---

# 103. Baseline Models

Every important model should have a baseline where practical.

Examples:

```text
Historical Average
Rule-Based Ranking
Simple ETA Formula
Moving Average Forecast
```

---

# 104. Baseline Comparison

A complex model must demonstrate meaningful improvement over a reasonable baseline.

---

# 105. AI vs Deterministic Algorithm

AI should not be introduced merely because it is technically interesting.

Use AI when it provides measurable value over a deterministic approach.

---

# 106. Model Selection Criteria

Select models based on:

```text
Accuracy
Latency
Cost
Reliability
Interpretability
Operational Complexity
Data Requirements
```

---

# 107. Model Complexity

Prefer the simplest model that meets the required business objective.

---

# 108. Over-Engineering Prevention

Do not introduce:

```text
Deep Learning
LLMs
Complex Ensembles
Real-Time Feature Infrastructure
```

unless simpler methods cannot satisfy the requirement.

---

# 109. Model Governance for Early Platform Stage

RideForge should initially prioritize:

```text
Reliable Data
Strong Baselines
Deterministic Constraints
Simple Models
Clear Evaluation
Operational Safety
```

before highly complex AI systems.

---

# 110. AI Experimentation

Experiments must be:

```text
Controlled
Measurable
Reversible
Documented
```

---

# 111. Experiment Metadata

Record:

```text
Experiment ID
Model Version
Variant
Traffic Allocation
Start Time
End Time
Metrics
Decision
```

---

# 112. Experiment Safety

Experiments must not intentionally violate:

```text
Legal Rules
Safety Rules
Security Rules
Hard Business Constraints
```

---

# 113. Experiment Rollback

Experiments must be disabled when:

```text
Critical Metrics Degrade
Safety Concern Appears
Unexpected Cost Occurs
Operational Instability Appears
```

---

# 114. Fairness

Models influencing driver or customer outcomes should be evaluated for unreasonable systematic bias where relevant.

---

# 115. Driver Fairness

Matching and ranking should consider whether model behaviour systematically disadvantages groups of drivers without legitimate operational justification.

---

# 116. Customer Fairness

Customer-facing predictions and decisions should be evaluated for unreasonable systematic disparities where applicable.

---

# 117. Fairness Does Not Override Constraints

Fairness evaluation must coexist with:

```text
Safety
Legal Requirements
Driver Eligibility
Operational Constraints
```

---

# 118. Human Oversight

High-impact AI systems should have an operational owner capable of:

```text
Reviewing
Disabling
Rolling Back
Escalating
```

the model.

---

# 119. Kill Switch

Critical AI capabilities should support rapid disablement.

Example:

```text
AI_RANKING_ENABLED=false
```

The exact mechanism should use the approved configuration and feature-flag infrastructure.

---

# 120. Kill Switch Safety

Disabling AI must result in:

```text
Approved Deterministic Fallback
```

rather than system failure.

---

# 121. AI Incident Response

AI incidents should follow:

```text
Detect
Contain
Disable / Rollback
Investigate
Correct
Evaluate
Document
```

---

# 122. AI Incident Examples

Examples:

```text
Unexpected Driver Ranking
ETA Regression
Demand Forecast Failure
Model Output Corruption
Feature Pipeline Failure
Provider AI Outage
Unexpected Cost Spike
```

---

# 123. AI Incident Evidence

Preserve sufficient information to determine:

```text
Model Version
Feature Version
Input Context
Output
Deployment
Configuration
```

without retaining unnecessary sensitive data.

---

# 124. Model Security

Protect models against:

```text
Unauthorized Replacement
Unauthorized Access
Artifact Tampering
Data Poisoning
Inference Abuse
```

---

# 125. Training Data Poisoning

Training pipelines should validate data sources and detect anomalous or corrupted training data where practical.

---

# 126. Model Artifact Tampering

Model artifacts should have integrity verification where supported.

---

# 127. Inference Abuse

Inference APIs should have:

```text
Authentication
Authorization
Rate Limiting
Resource Limits
```

as appropriate.

---

# 128. AI Privacy

AI systems must follow:

```text
ADR-0023 — Security and Secret Management
```

and the platform's privacy requirements.

---

# 129. Data Minimization for Inference

Send only the data required to produce the prediction.

---

# 130. Sensitive Feature Restrictions

Sensitive attributes should not be used as model features unless explicitly justified, governed, and legally permitted.

---

# 131. Model Documentation

Every production model should have a model card or equivalent documentation.

Minimum content:

```text
Purpose
Owner
Version
Training Data
Features
Evaluation
Known Limitations
Failure Modes
Fairness Considerations
Deployment
Rollback
```

---

# 132. Model Limitations

Known limitations must be documented.

Example:

```text
Poor Performance in Sparse Regions
Poor Cold-Start Performance
Reduced Accuracy During Unusual Events
```

---

# 133. Model Approval Checklist

Before production:

```text
□ Purpose Defined
□ Owner Assigned
□ Dataset Versioned
□ Features Versioned
□ Baseline Established
□ Offline Evaluation Completed
□ Business Metrics Evaluated
□ Safety Constraints Verified
□ Privacy Review Completed
□ Security Review Completed
□ Latency Tested
□ Cost Evaluated
□ Fallback Implemented
□ Rollback Identified
□ Monitoring Implemented
□ Approval Recorded
```

---

# 134. Production Monitoring Checklist

After deployment:

```text
□ Latency
□ Error Rate
□ Timeout Rate
□ Prediction Distribution
□ Feature Availability
□ Feature Drift
□ Model Drift
□ Business Metrics
□ Fallback Rate
□ Cost
□ Safety Signals
```

---

# 135. Model Deployment Checklist

```text
□ Approved Model Version
□ Correct Artifact
□ Correct Feature Version
□ Correct Configuration
□ Correct Environment
□ Monitoring Enabled
□ Rollback Ready
□ Canary / Shadow Strategy
□ Owner Available
```

---

# 136. Model Retirement Checklist

```text
□ Replacement Available
□ Consumers Migrated
□ Production Traffic Removed
□ Model Marked Retired
□ Metadata Preserved
□ Artifact Retention Applied
□ Documentation Updated
```

---

# 137. AI Architecture Boundary

AI should remain a replaceable component.

Conceptually:

```text
Ride Domain
    │
    ▼
Dispatch Application
    │
    ▼
AI Interface
    │
    ├── Model A
    ├── Model B
    └── Deterministic Fallback
```

The domain should not depend directly on a specific model implementation.

---

# 138. Model Abstraction

Use an application-level abstraction such as:

```text
RankCandidates(...)
PredictETA(...)
ForecastDemand(...)
```

rather than coupling business logic directly to a model runtime.

---

# 139. AI Provider Independence

Where practical, AI provider dependencies should remain replaceable.

Avoid embedding provider-specific assumptions throughout the domain.

---

# 140. Model Runtime Independence

The application should not require knowledge of:

```text
TensorFlow
PyTorch
ONNX
External AI API
```

unless that knowledge belongs in the AI infrastructure layer.

---

# 141. AI Failure Boundary

The AI boundary should return controlled results such as:

```text
Prediction
Confidence / Score
Model Version
Error / Timeout
```

rather than leaking infrastructure details into domain logic.

---

# 142. AI Observability Boundary

AI services should emit:

```text
Inference Metrics
Model Version
Latency
Errors
Fallback
```

through the platform observability system.

---

# 143. AI Configuration Boundary

AI configuration should follow:

```text
ADR-0024 — Configuration and Environment Strategy
```

---

# 144. AI Security Boundary

AI credentials and privileged access follow:

```text
ADR-0023 — Security and Secret Management
```

---

# 145. AI Testing Boundary

AI testing follows:

```text
ADR-0025 — Testing and Integration Strategy
```

---

# 146. AI Failure Strategy

AI fallback follows:

```text
ADR-0021 — Failure and Degradation Strategy
```

---

# 147. AI Architecture Strategy

AI architecture follows:

```text
05-ai/
```

documentation and the broader RideForge architecture.

---

# 148. Consequences

## 148.1 Positive Consequences

This governance model provides:

```text
Controlled AI Adoption
Model Traceability
Safer Production Deployment
Reproducibility
Better Rollback
Clear Ownership
Improved Incident Response
Reduced AI Blast Radius
```

---

## 148.2 Negative Consequences

The model introduces:

```text
Model Registry Work
Evaluation Infrastructure
Monitoring Requirements
Data Governance
Approval Processes
Model Documentation
Experiment Management
Additional Operational Complexity
```

These costs are accepted for production AI systems.

---

# 149. Risks

## Risk 1 — AI Becomes Too Central

### Mitigation

Maintain deterministic fallbacks and keep domain rules outside models.

---

## Risk 2 — Model Drift

### Mitigation

Monitor:

```text
Feature Drift
Prediction Drift
Performance Drift
```

and retrain or rollback when required.

---

## Risk 3 — Training Data Leakage

### Mitigation

Use temporal validation and strict dataset construction.

---

## Risk 4 — Model Changes Without Traceability

### Mitigation

Immutable model versions and registry metadata.

---

## Risk 5 — AI Cost Explosion

### Mitigation

Monitor:

```text
Inference Volume
Compute
Provider Usage
Latency
```

and apply budgets and limits.

---

## Risk 6 — Unsafe Model Output

### Mitigation

Validate model output and enforce hard constraints outside the model.

---

## Risk 7 — Vendor Lock-In

### Mitigation

Use application-level AI interfaces and isolate provider-specific integrations.

---

## Risk 8 — Excessive AI Complexity

### Mitigation

Prefer the simplest model that provides measurable business value.

---

# 150. Alternatives Considered

## 150.1 Let AI Control Final Decisions

### Advantages

```text
Maximum Model Flexibility
Potentially Strong Optimization
```

### Disadvantages

```text
Safety Risk
Legal Risk
Poor Explainability
Large Blast Radius
```

### Decision

```text
Rejected.
```

---

# 151. No Model Registry

### Advantages

```text
Simple
```

### Disadvantages

```text
Poor Traceability
Difficult Rollback
Difficult Reproducibility
```

### Decision

```text
Rejected for production models.
```

---

# 152. Deploy Latest Model Automatically

### Advantages

```text
Fast Iteration
```

### Disadvantages

```text
Uncontrolled Behaviour
Regression Risk
Operational Risk
```

### Decision

```text
Rejected.
```

---

# 153. Retrain Automatically and Deploy Automatically

### Advantages

```text
Low Manual Effort
```

### Disadvantages

```text
Unvalidated Models
Data Quality Risk
Drift Amplification
Unexpected Behaviour
```

### Decision

```text
Rejected as the default strategy.
```

---

# 154. Use AI Everywhere

### Advantages

```text
Maximum AI Utilization
```

### Disadvantages

```text
Complexity
Cost
Latency
Unnecessary Risk
```

### Decision

```text
Rejected.
```

---

# 155. Validation

This ADR should be validated through:

```text
Model Registry Validation
Training Pipeline Validation
Evaluation Tests
Feature Validation
Model Serving Tests
Shadow Deployment
Canary Deployment
Rollback Tests
AI Failure Tests
Security Tests
Privacy Tests
Fairness Evaluation
Cost Monitoring
Production Monitoring
```

---

# 156. Review Triggers

Revisit this ADR when:

```text
A New Production Model Is Introduced
A New AI Provider Is Added
AI Gains New Operational Authority
A New Region Is Added
A Model Causes a Production Incident
Training Data Sources Change
Privacy Requirements Change
Model Serving Architecture Changes
AI Cost Becomes Material
A Major Model Architecture Changes
Regulatory Requirements Change
```

---

# 157. Final Principles

The following principles are mandatory:

```text
1. AI is an optimization and prediction capability, not the ultimate authority over platform rules.

2. Hard business, legal, security, and safety constraints remain deterministic.

3. AI must not bypass regional or legal ride validation.

4. AI must not bypass authorization.

5. AI must not bypass safety constraints.

6. Production models must have immutable versions.

7. Every production model must have traceable metadata.

8. Training datasets must be versioned where material changes occur.

9. Feature definitions must be versioned where their semantics change.

10. Training and serving feature definitions must remain consistent where they represent the same feature.

11. Temporal leakage must be prevented.

12. Models must be evaluated before production deployment.

13. Model evaluation must include both technical and business metrics.

14. A reasonable deterministic baseline should be used where practical.

15. Model complexity must be justified by measurable value.

16. Production model deployment must be controlled.

17. New models should use shadow, canary, or similarly controlled rollout mechanisms where appropriate.

18. Every production model must have a rollback path.

19. AI failure must not become system failure when a safe fallback exists.

20. Model outputs must be validated before affecting business decisions.

21. Invalid model outputs must be rejected.

22. AI must not directly perform privileged actions without deterministic validation.

23. Model inference must have bounded latency.

24. Model inference must have bounded resource usage.

25. Model and feature drift must be monitored.

26. Model performance must be monitored against real outcomes where ground truth is available.

27. Retraining does not automatically imply production deployment.

28. Retrained models require evaluation and approval.

29. Model artifacts must be protected from unauthorized modification.

30. Production model deployment permissions must be controlled.

31. Model development and production deployment should have separation of duties where practical.

32. AI provider credentials must follow secret-management requirements.

33. External AI providers must receive only the data required for the task.

34. Sensitive data must not be sent to AI providers without appropriate governance.

35. Production prompts that materially affect behaviour must be versioned and reviewed.

36. AI experiments must be controlled, measurable, and reversible.

37. AI experiments must not bypass hard constraints.

38. AI systems influencing customer or driver outcomes should be evaluated for unreasonable systematic bias where relevant.

39. AI systems must have operational ownership.

40. Critical AI capabilities should have a kill switch.

41. Disabling AI must activate an approved fallback.

42. AI incidents must be investigated with sufficient model and deployment context.

43. Historical model metadata must remain available according to retention requirements.

44. AI architecture should isolate model implementations from core domain logic.

45. The domain should depend on AI capabilities through stable application interfaces.

46. AI provider-specific implementation should remain outside the domain layer.

47. Model governance must remain aligned with security, configuration, testing, observability, and failure-management ADRs.

48. AI should be introduced only where it provides measurable value over a simpler alternative.

49. The safest effective model is preferred over the most complex model.

50. The AI system must remain controllable, observable, reversible, and bounded by deterministic platform rules.
```

---

# 158. Status

```text
Decision: ACCEPTED

Governance Model:
Controlled + Versioned + Evaluated + Observable + Reversible

AI Authority:
Optimizer / Predictor / Ranker

Hard Constraints:
Deterministic

Legal Constraints:
Outside Model Authority

Security Constraints:
Outside Model Authority

Model Registry:
Required for Production Models

Model Versioning:
Immutable Production Versions

Training Data:
Versioned and Governed

Features:
Defined + Versioned + Validated

Evaluation:
Technical + Business Metrics

Baseline:
Required Where Practical

Deployment:
Controlled

Rollout:
Shadow / Canary / Controlled Experiment Where Appropriate

Monitoring:
Required

Drift Monitoring:
Required Where Ground Truth / Distribution Monitoring Is Available

Rollback:
Required

Fallback:
Required for Critical AI Paths

Model Output:
Validated

AI Security:
ADR-0023

Configuration:
ADR-0024

Testing:
ADR-0025

Failure Strategy:
ADR-0021

Observability:
ADR-0022

Dispatch:
Smart + Stand + Hybrid

Primary Goal:
Use AI to Improve RideForge While Keeping the Platform Deterministic, Safe, Governable, Observable, and Operationally Recoverable
```

---

# 159. Decision Summary

RideForge adopts the following AI control model:

```text
                    RIDE REQUEST
                         │
                         ▼
                ┌─────────────────┐
                │ Hard Constraints│
                │ Legal / Safety  │
                │ Authorization   │
                └────────┬────────┘
                         │
                         ▼
                 ELIGIBLE SET
                         │
                         ▼
                ┌─────────────────┐
                │ Feature Layer   │
                └────────┬────────┘
                         │
                         ▼
                ┌─────────────────┐
                │ AI / ML Model   │
                │ Ranking / ETA   │
                │ Prediction      │
                └────────┬────────┘
                         │
                         ▼
                OUTPUT VALIDATION
                         │
                  ┌──────┴──────┐
                  │             │
                Valid         Invalid
                  │             │
                  ▼             ▼
             Business       Fallback
             Decision       Strategy
                  │             │
                  └──────┬──────┘
                         ▼
                    FINAL ACTION
```

The key architectural boundary is:

```text
AI MAY OPTIMIZE
```

but:

```text
AI MAY NOT OVERRIDE
```

the deterministic rules that keep RideForge:

```text
Legal
Secure
Safe
Consistent
Operational
Recoverable
```

---

# 160. Status Metadata

| Field | Value |
|---|---|
| ADR | `0026` |
| Title | Model and AI Governance |
| Status | Accepted |
| Category | AI / ML / Governance |
| AI Authority | Bounded Optimization / Prediction |
| Hard Constraints | Deterministic |
| Model Registry | Required |
| Model Versioning | Immutable Production Versions |
| Training Data | Versioned / Governed |
| Feature Governance | Required |
| Model Evaluation | Required |
| Baseline Comparison | Required Where Practical |
| Controlled Rollout | Recommended for Material Changes |
| Monitoring | Required |
| Drift Monitoring | Required Where Applicable |
| Rollback | Required |
| Fallback | Required for Critical AI Paths |
| AI Output Validation | Required |
| Security | ADR-0023 |
| Configuration | ADR-0024 |
| Testing | ADR-0025 |
| Failure Strategy | ADR-0021 |
| Observability | ADR-0022 |
| Next ADR | `0027-CLOUD_AND_DEPLOYMENT_STRATEGY.md` |
