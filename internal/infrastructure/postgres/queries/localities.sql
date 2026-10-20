-- name: LocalityCellsForBoundary :one
SELECT ST_AsGeoJSON(geom)
		FROM (SELECT ST_Multi(ST_GeomFromText($1, 4326)) AS geom) AS boundary
		WHERE GeometryType(geom) = 'MULTIPOLYGON'
		  AND ST_IsValid(geom)
		  AND NOT ST_IsEmpty(geom);

-- name: LocalityCreate :execrows
INSERT INTO localities (id, administrative_region_id, name, slug, boundary, status)
		VALUES ($1, $2, $3, $4, ST_Multi(ST_GeomFromText($5, 4326)), $6);

-- name: LocalityUpdateBoundary :execrows
UPDATE localities
		SET boundary = ST_Multi(ST_GeomFromText($1, 4326)), updated_at = NOW()
		WHERE id = $2 AND ST_IsValid(ST_Multi(ST_GeomFromText($1, 4326)))
		  AND NOT ST_IsEmpty(ST_Multi(ST_GeomFromText($1, 4326)));

-- name: LocalityDeleteH3Cells :execrows
DELETE FROM locality_h3_cells WHERE locality_id = $1;

-- name: LocalityInsertH3Cells :execrows
INSERT INTO locality_h3_cells (locality_id, h3_cell, h3_resolution)
		SELECT $1, cell, $3 FROM unnest($2::text[]) AS cell;

-- name: LocalityUpdateStatus :execrows
UPDATE localities SET status = $1, updated_at = NOW() WHERE id = $2;

-- name: LocalityActivate :execrows
UPDATE localities SET status = $1, updated_at = NOW() WHERE id = $2 AND EXISTS (SELECT 1 FROM locality_h3_cells WHERE locality_id = $2 AND h3_resolution = $3);

-- name: LocalityResolveH3Cell :many
SELECT l.id
		FROM locality_h3_cells c
		JOIN localities l ON l.id = c.locality_id
		WHERE c.h3_cell = $1 AND c.h3_resolution = $2 AND l.status = 'ACTIVE'
		LIMIT 2;
