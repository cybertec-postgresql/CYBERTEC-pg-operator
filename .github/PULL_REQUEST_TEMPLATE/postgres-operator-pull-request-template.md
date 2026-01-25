## Description
## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update


## Linked Issues
---

## Checklist

Thanks for submitting a pull request to the CYBERTEC-pg-operator project!
Please ensure your contribution matches the following items:

- [ ] **Code Formatting:** My code follows the Go formatting standards. (Your IDE should do this automatically, or run `go fmt`).
- [ ] **Generated Code:** I have updated [generated code](https://cybertec-postgresql.github.io/CYBERTEC-pg-operator/contribution#code-generation) (clientset, deepcopy) when introducing new fields to the `cpo.opensource.cybertec.at` API package.
- [ ] **Configuration & CRDs:** New [configuration options](https://cybertec-postgresql.github.io/CYBERTEC-pg-operator/contribution#introduce-additional-configuration-parameters) are reflected in:
    - [ ] The Go struct definitions
    - [ ] The CRD validation manifests (YAML)
    - [ ] The Helm Charts
    - [ ] The sample manifests
- [ ] **Tests:** New functionality is covered by [unit tests](https://cybertec-postgresql.github.io/CYBERTEC-pg-operator/contribution#unit-tests) and/or [e2e tests](https://cybertec-postgresql.github.io/CYBERTEC-pg-operator/contribution#end-to-end-tests).
- [ ] **Existing PRs:** I have checked existing open PRs to ensure there are no duplicates or conflicts.

## How has this been tested?