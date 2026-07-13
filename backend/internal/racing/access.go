package racing

import "context"

// Access describes how a user can run a piece of content.
type Access string

const (
	AccessFree    Access = "free"    // default content — everyone has it
	AccessOwned   Access = "owned"   // purchased (directly, via sku group, or via requirements)
	AccessMissing Access = "missing" // not available to this user
)

// trackAccess resolves every track's access for a user:
//   - free content is available to everyone;
//   - a purchase unlocks every config sharing the same sku_group;
//   - a combined layout with track_requirements is owned when ALL its required
//     tracks are available (free or owned).
func (s *Service) trackAccess(ctx context.Context, userID int64) (map[int]Access, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT t.track_id, t.is_free, t.sku_group, o.track_id IS NOT NULL AS owned
		FROM tracks t
		LEFT JOIN owned_tracks o ON o.track_id = t.track_id AND o.user_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	access := make(map[int]Access)
	skuOf := make(map[int]int)
	ownedSkus := make(map[int]bool)
	for rows.Next() {
		var (
			id, sku       int
			free, ownedTr bool
		)
		if err := rows.Scan(&id, &free, &sku, &ownedTr); err != nil {
			return nil, err
		}
		skuOf[id] = sku
		switch {
		case free:
			access[id] = AccessFree
		case ownedTr:
			access[id] = AccessOwned
			if sku != 0 {
				ownedSkus[sku] = true
			}
		default:
			access[id] = AccessMissing
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// One purchase unlocks all configs in the same sku group.
	for id, sku := range skuOf {
		if access[id] == AccessMissing && sku != 0 && ownedSkus[sku] {
			access[id] = AccessOwned
		}
	}

	// Combined layouts: owned when every required track is available.
	reqs, err := s.loadRequirements(ctx)
	if err != nil {
		return nil, err
	}
	for trackID, required := range reqs {
		if access[trackID] != AccessMissing || len(required) == 0 {
			continue
		}
		all := true
		for _, r := range required {
			if a, ok := access[r]; !ok || a == AccessMissing {
				all = false
				break
			}
		}
		if all {
			access[trackID] = AccessOwned
		}
	}
	return access, nil
}

func (s *Service) loadRequirements(ctx context.Context) (map[int][]int, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT track_id, requires_track_id FROM track_requirements`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reqs := make(map[int][]int)
	for rows.Next() {
		var track, req int
		if err := rows.Scan(&track, &req); err != nil {
			return nil, err
		}
		reqs[track] = append(reqs[track], req)
	}
	return reqs, rows.Err()
}
