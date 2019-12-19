# drone-vault-reposecrets

A drone secret plugin that reads secrets based on a fixed path and repo name.

In contrast to the official vault extension this one expects a secret per repo. If you need access to multiple secrets, consider using the [original plugin][1] .

## Envionment

* `PLUGIN_SECRET`: plugin secret to communicate with drone
* `PLUGIN_LISTEN`: http listen address

* `VAULT_SECRET_PATH`: path to the vault secrets, e.g. `kv/drone/%s`, `%s` is replaces by the repo slug (**required**)

* `VAULT_ADDR`: vaults url (**required**)
* `VAULT_ROLE_ID`: role id for vault login (**required**)
* `VAULT_SECRET_ID`: secret id for vault login (**required**)


[1]: https://github.com/drone/drone-vault
