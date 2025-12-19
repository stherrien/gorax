function(ctx) {
  identity_id: ctx.identity.id,
  email: ctx.identity.traits.email,
  name: if std.objectHas(ctx.identity.traits, 'name') then ctx.identity.traits.name else {},
  tenant_id: if std.objectHas(ctx.identity.traits, 'tenant_id') then ctx.identity.traits.tenant_id else null,
  created_at: ctx.identity.created_at,
  updated_at: ctx.identity.updated_at,
}
