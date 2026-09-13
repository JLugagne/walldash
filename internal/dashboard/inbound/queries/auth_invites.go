package queries

import "net/http"

// Invites lists the invitations (newest first) without their plaintext tokens, which are only
// ever returned once at creation time.
func (h *AuthHandler) Invites(w http.ResponseWriter, r *http.Request) {
	all, err := h.auth.ListInvites(r.Context())
	if err != nil {
		h.controller.SendError(w, r, err)
		return
	}
	data := make([]map[string]any, 0, len(all))
	for _, inv := range all {
		data = append(data, map[string]any{
			"selector":    inv.Selector,
			"role":        inv.Role,
			"created_by":  inv.CreatedBy,
			"created_at":  inv.CreatedAt,
			"expires_at":  inv.ExpiresAt,
			"consumed_at": inv.ConsumedAt,
			"consumed_by": inv.ConsumedBy,
			"revoked_at":  inv.RevokedAt,
		})
	}
	h.controller.SendSuccess(w, r, data)
}
