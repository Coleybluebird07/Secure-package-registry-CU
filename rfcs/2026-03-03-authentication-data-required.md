+++
title = "The data required for proper authentication and identification"
authors = ["readb5@cardiff.ac.uk"]
+++

## Summary

Discussing the data required for proper authentication and identification when pulling packages from the registry, in
order to enforce our proposed pricing models and to ensure that we are only permitting users and organisations to pull
packages that they are authorized to pull.

The proposal is currently to have a minimal schema addition, with organisations, users and a link table between
organisations and packages to track what packages certain organisations are authorized to pull.

## Problem

To properly track what packages organizations and individuals are pulling on the registry, and to ensure that we are
only permitting them to pull packages that they are authorized to pull, we need to have a clear understanding of the
data that is required for proper authentication and identification. This is also useful because it is essential for
understanding what units we can price packages around (ie: per organisation, per user, per package, etc).

This breaks down the data users will need to provide in order to authenticate and identify themselves when pulling
packages from the registry, and the data that organizations will need to provide in order to authenticate and identify
themselves when pulling packages from the registry.

## Recommendation

What we need to track:

- The organisation
- The users (and their organisation)
- The packages being pulled
- The packages that certain users and organisations are authorized to pull

This is based on the potential pricing models we have discussed through our business case report last semester. I
propose this simple schema to start with:

### Rough Database Schema

See core-data-schema.md for all the package data

```sql

CREATE TABLE organisations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
);

--Link table to avoid many-to-many relationship between organisations and packages, since an organisation can have many packages, and a package can be used by many organisations.
CREATE TABLE orgnaisation_packages (
    id SERIAL PRIMARY KEY,
    organisation_id INTEGER REFERENCES organisations(id),
    package_id REFERENCES packages(id)
);

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    organisation_id INTEGER REFERENCES organisations(id)
);

```

With this schema, we can track what packages organisations and their users are pulling, the volume of packages being
pulled, and we can also track what packages certain users and organisations are authorized to pull. This will allow us
to implement a pricing model based on the number of packages being pulled, the number of users, or the number of
organisations.

## Open Questions

I've opted for a very simply database schema here, one that will likely need to be extended in the future to include
more information about the organisations and users (such as the team the user is on, and their role, since different
users will need different permissions). Should we try for a more ambitious database schema now, or should we start with
something simple and iterate on it as we go?
