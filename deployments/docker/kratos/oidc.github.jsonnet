local claims = std.extVar('claims');

{
  identity: {
    traits: {
      [if 'email' in claims then 'email' else null]: claims.email,
      [if 'name' in claims then 'name' else null]: {
        first: if 'name' in claims then std.split(claims.name, ' ')[0] else '',
        last: if 'name' in claims && std.length(std.split(claims.name, ' ')) > 1 then std.split(claims.name, ' ')[1] else '',
      },
    },
  },
}
