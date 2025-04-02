---
title: "Databases & Users"
date: 2023-12-28T14:26:51+01:00
draft: false
weight: 900
---

CPO not only supports you in deploying your cluster, it also supports you in setting it up in terms of the database and users. 
CPO offers you three different options for this: 
- Create roles
- Create databases
- preapared databases

## Create Roles
The creation of users is based on the definition of the user name and the definition of the required rights for this user. Available rights are
- `superuser`
- `inherit`
- `login`
- `nologin`
- `createrole`
- `createdb`
- `replication`
- `bypassrls`

Unless explicitly defined via `NOLOGIN`, a created user automatically receives the `LOGIN` permission. 

```
spec:
  users:
    db_owner:
    - login
    - createdb
    appl_user:
    - login
```

For each user created, CPO automatically creates a secret with `username` and `password` in the namespace of the cluster, which follows the following naming convention: 
[USERNAME].[CLUSTERNAME].credentials.postgresql.cpo.opensource.cybertec.at 

If the secrets for an application are to be stored in a different namespace, for example, it is necessary to define the setting enable_cross_namespace_secret as true in the operator configuration. You can find more information about the operator configuration [here](documentation/how-to-use/operator_configuration/).

The namespace must then be written before the user name.
```
spec:
  users:
    db_owner:
    - login
    - createdb
    app_namespace.appl_user:
    - login
```

## Create Databases 

Databases are basically created in a very similar way to users.
The definition is based on the database name and the database owner. 

```
spec:
  users:
    db_owner:
    - login
    - createdb
    app_namespace.appl_user:
    - login
  databases;
    app_db: app_namespace.appl_user
```

{{< hint type=Info >}}Be aware that the user name must be defined for the database owner in the same way as it is done in the users object. {{< /hint >}}

## Prepared Databases

The `preparedDatabases` object is available for a much more extensive setup of databases and users. 
In addition to the creation of `databases` and `users`, this also enables the creation of `schemas` and `extensions`. A more detailed rights management is also available.

### Databases and Schema

Creating the preparedDatabases object already creates a database whose name is based on the cluster name. `preparedDatabases: {}`

{{< hint type=Info >}}For the database name, `-` is replaced with `_` in the cluster name{{< /hint >}}

To create your own database names and elements such as schemas and extensions within the database, an object must be created within preparedDatabases for each database.

```
spec:
  preparedDatabases:
    appl_db:
      extensions:
        dblink: public
      schemas:
        data: {}
```

This example creates a database with the name `appl_db` and creates a schema with the name `data` in it, as well as creating the `dblink` extension in the schema `public`.

### Management of users and Permissions

For rights management, we distinguish between `NOLOGIN` roles and `LOGIN` roles. `Users` have login rights and inherit the other rights from the `NOLOGIN` role. 

#### NoLogin roles (defaultRoles)

The roles are created if `defaultroles` is not explicitly set to false.
```
spec:
  preparedDatabases:
    appl_db:
      extensions:
        dblink: public
      schemas:
        data: {}
```
This creates roles for the schema owner, writer and reader

#### Login roles (defaultUsers)

The roles described in the previous paragraph can be assigned to LOGIN roles via the users section in the manifest. Optionally, the Postgres operator can also create standard `LOGIN` roles for the database and each individual schema. These roles are given the suffix _user and inherit all rights from their NOLOGIN counterparts. Therefore, you cannot set defaultRoles to false and activate defaultUsers at the same time.

```
spec:
  preparedDatabases:
    appl_db:
      defaultUsers: true
      extensions:
        dblink: public
      schemas:
        data: {}
        history:
          defaultRoles: true
          defaultUsers: false
```
This example creates the following users and inheritances

Role name               | Attributes                | inherits from
------------------------|---------------------------|--------------------------------
 appl_db_owner          | Cannot login              | appl_db_reader,appl_db_owner,appl_data_owner,...
 appl_db_owner_user     |                           | appl_db_owner
 appl_db_reader         | Cannot login              |
 appl_db_reader_user    |                           | appl_db_reader
 appl_db_writer         | Cannot login              | appl_db_reader
 appl_db_writer_user    |                           | appl_db_writer
 appl_db_data_owner     | Cannot login              | appl_db_data_reader,appl_db_data_writer
 appl_db_data_reader    | Cannot login              |
 appl_db_data_writer    | Cannot login              | appl_db_data_reader
 appl_db_history_owner  | Cannot login              | appl_db_history_reader,appl_db_history_writer
 appl_db_history_reader | Cannot login              |
 appl_db_history_writer | Cannot login              | appl_db_history_reader

Default access permissions are also defined for LOGIN roles when databases and schemas are created. This means that they are not currently set if defaultUsers (or defaultRoles for schemas) are activated at a later time.

#### User Secrets

For each user created by cpo with `LOGIN` permissions, the operator also creates a secret with username and password, as with the creation of roles via the `users` object.