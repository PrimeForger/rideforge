-- name: SurgePolicyLoadActive :many
SELECT p.id, p.locality_id, p.version, p.status,
		       p.min_multiplier, p.max_multiplier, p.minimum_demand,
		       p.effective_from, p.effective_until,
		       t.min_imbalance_ratio, t.multiplier
		FROM surge_policies p
	LEFT JOIN surge_policy_tiers t ON t.policy_id = p.id
		WHERE p.status = 'ACTIVE'
		  AND p.effective_from <= $1
		  AND (p.effective_until IS NULL OR p.effective_until > $1)
		ORDER BY p.locality_id, p.id, t.min_imbalance_ratio;
