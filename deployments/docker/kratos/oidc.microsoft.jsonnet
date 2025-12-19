local claims = std.extVar('claims');

{
  identity: {
    traits: {
      [if 'email' in claims then 'email' else null]: claims.email,
      [if 'name' in claims then 'name' else null]: {
        first: if 'given_name' in claims then claims.given_name else '',
        last: if 'family_name' in claims then claims.family_name else '',
      },
    },
  },
}
