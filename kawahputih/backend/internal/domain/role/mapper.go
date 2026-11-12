package role

func toRoleResponse(r Role) RoleResponse {
	perms := make([]string, 0, len(r.Permissions))
	for _, p := range r.Permissions {
		perms = append(perms, p.Code)
	}
	return RoleResponse{ID: r.ID, Name: r.Name, Description: r.Description, Permissions: perms}
}

func toRoleResponseList(roles []Role) []RoleResponse {
	out := make([]RoleResponse, 0, len(roles))
	for _, r := range roles {
		out = append(out, toRoleResponse(r))
	}
	return out
}
