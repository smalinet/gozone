
## [v0.5.3] - 2026-06-05

### 💼 Other

- **(handlers)** ([0f49276](https://github.com/babykart/gozone/commit/0f49276916e8812c6426d0dc134f0abc8b4aa601)) - [gozone] log internal errors server-side instead of exposing them to users - ([smalinet](https://github.com/smalinet))
- ([b3d7824](https://github.com/babykart/gozone/commit/b3d78245620518b26a2be5495963b536bb59976b)) - Merge pull request #4 from smalinet/fix/batch-create-logs-order

fix(handlers): [gozone] log activity after CreateRecords succeeds in … - ([babykart](https://github.com/babykart))
- ([7e03412](https://github.com/babykart/gozone/commit/7e034129cbc09fd4eba1a141e9376165398dbbdc)) - Merge pull request #5 from smalinet/fix/activity-log-username

fix(handlers): [gozone] populate Username field in zone activity logs - ([babykart](https://github.com/babykart))
- ([7e15ffb](https://github.com/babykart/gozone/commit/7e15ffb2e620dd5deead5270113458f37d8f67b1)) - Merge pull request #6 from smalinet/fix/validator-reverse-dns

fix(validators): [gozone] allow digit-leading labels to support rever… - ([babykart](https://github.com/babykart))
- ([e2da294](https://github.com/babykart/gozone/commit/e2da2946d92e2cf1bfa48a8f37969325655ee0ce)) - Merge pull request #7 from smalinet/fix/hide-internal-errors

security(handlers): [gozone] log internal errors server-side instead … - ([babykart](https://github.com/babykart))
- ([1826fd9](https://github.com/babykart/gozone/commit/1826fd92189f82790f31cb6ba2fbcf0127f7bfe2)) - Merge pull request #8 from smalinet/fix/zones-n-plus-one

perf(pdns): [gozone] eliminate N+1 HTTP requests on zone list page - ([babykart](https://github.com/babykart))
- ([4810a5a](https://github.com/babykart/gozone/commit/4810a5af53331237211761622e92f81e9b4d769c)) - Merge pull request #15 from smalinet/fix/remove-duplicate-startup-log

fix(cmd): [gozone] remove duplicate startup log line - ([babykart](https://github.com/babykart))
- ([c4deaa9](https://github.com/babykart/gozone/commit/c4deaa951901b1b2015dc725c799e50836faea2f)) - Merge branch main into fix/pdns-bounded-goroutines

Signed-off-by: smalinet <s.malinet@monaco-telecom.mc> - ([smalinet](https://github.com/smalinet))
- ([90a6a2e](https://github.com/babykart/gozone/commit/90a6a2e4a9bf054342835cc4ee94b335c9e42188)) - Merge pull request #9 from smalinet/fix/pdns-bounded-goroutines

fix(pdns): [gozone] cap concurrent goroutines in ListZonesWithInfo wi… - ([babykart](https://github.com/babykart))
- ([ca4b369](https://github.com/babykart/gozone/commit/ca4b36914d61ba287d2f2b2cf0ed50ef23c3bc4a)) - Merge branch main into fix/view-zone-parallel-calls

Signed-off-by: smalinet <s.malinet@monaco-telecom.mc> - ([smalinet](https://github.com/smalinet))
- ([8da03bd](https://github.com/babykart/gozone/commit/8da03bd36885a999452dd7e2210b572b1c94823c)) - Merge pull request #10 from smalinet/fix/view-zone-parallel-calls

perf(handlers): [gozone] parallelize independent PowerDNS calls in ViewZone - ([babykart](https://github.com/babykart))
- ([f0126a4](https://github.com/babykart/gozone/commit/f0126a4f42756d1e23acfc97bb9b861049ade6b1)) - Merge branch main into fix/pdns-context-propagation

Signed-off-by: smalinet <s.malinet@monaco-telecom.mc> - ([smalinet](https://github.com/smalinet))
- ([202918e](https://github.com/babykart/gozone/commit/202918ea590b616d24d784c02a085357a4ca8256)) - Merge pull request #11 from smalinet/fix/pdns-context-propagation

fix(pdns): [gozone] propagate request context to all PowerDNS HTTP calls - ([babykart](https://github.com/babykart))
- ([31857e8](https://github.com/babykart/gozone/commit/31857e863f3589ef6b72241f00eebb765978d14a)) - Merge pull request #12 from smalinet/fix/remove-redundant-method-guards

refactor(handlers): [gozone] remove redundant r.Method != http.Method - ([babykart](https://github.com/babykart))
- ([686edd4](https://github.com/babykart/gozone/commit/686edd45966b228d39e0fec2b1e1e73f881fcacb)) - Merge pull request #13 from smalinet/fix/url-normalization

refactor(pdns): [gozone] simplify URL normalization in NewClient - ([babykart](https://github.com/babykart))
- ([5c5d02a](https://github.com/babykart/gozone/commit/5c5d02a108accde0fed51f5ce7e68bb63e794d08)) - Merge pull request #14 from smalinet/fix/remove-double-getuser

refactor(handlers): [gozone] eliminate duplicate GetUser call in render - ([babykart](https://github.com/babykart))

### 🐛 Bug Fixes

- **(cmd)** ([434c1c7](https://github.com/babykart/gozone/commit/434c1c73e40b4170a2f1619d866679e8f41ded7b)) - [gozone] remove duplicate startup log line - ([smalinet](https://github.com/smalinet))
- **(cmd)** ([18f1fef](https://github.com/babykart/gozone/commit/18f1feface04f4d4c208c8f8e291177d3241ca77)) - [gozone] fix MIME type for the *.ico - ([babykart](https://github.com/babykart))
- **(handlers)** ([b613f6a](https://github.com/babykart/gozone/commit/b613f6a4cd062d86ae7fc4f1a63f9c9af3b4d9b8)) - [gozone] log activity after CreateRecords succeeds in BatchCreateRecords - ([smalinet](https://github.com/smalinet))
- **(handlers)** ([763a839](https://github.com/babykart/gozone/commit/763a839390161579c9e9c631181ca5eb310a28bf)) - [gozone] populate Username field in zone activity logs - ([smalinet](https://github.com/smalinet))
- **(handlers)** ([4b2cbfa](https://github.com/babykart/gozone/commit/4b2cbfa032c4ba67d32270668f9318604165fe9a)) - [gozone] fix empty user & zones in group add - ([babykart](https://github.com/babykart))
- **(pdns)** ([636ee1e](https://github.com/babykart/gozone/commit/636ee1e6e75e79dd5c3c14881a0e82852bdc8c3b)) - [gozone] propagate request context to all PowerDNS HTTP calls - ([smalinet](https://github.com/smalinet))
- **(pdns)** ([a92aa1c](https://github.com/babykart/gozone/commit/a92aa1ca68342db1d130ee9c5674c877f66db5af)) - [gozone] cap concurrent goroutines in ListZonesWithInfo with a semaphore - ([smalinet](https://github.com/smalinet))
- **(validators)** ([dd0070d](https://github.com/babykart/gozone/commit/dd0070d177d9b5e80ab6974419adad0b1ae3d44b)) - [gozone] allow digit-leading labels to support reverse DNS zones - ([smalinet](https://github.com/smalinet))
- **(web)** ([61d016a](https://github.com/babykart/gozone/commit/61d016ae52a762515fd0d54327cf95a13d7f93fe)) - [gozone] align stats in PowerDNS server card into a proper 2-column grid - ([babykart](https://github.com/babykart))
- **(web)** ([2572f0d](https://github.com/babykart/gozone/commit/2572f0d0fd0f049f655901410b0b82a29eda02b0)) - [gozone] compact table styling with reduced padding and font-size - ([babykart](https://github.com/babykart))

### 🚜 Refactor

- **(handlers)** ([3f08dbb](https://github.com/babykart/gozone/commit/3f08dbb7d3af85257a8b547d11c8f629ccd1b4ae)) - [gozone] remove redundant r.Method != http.MethodPost guards - ([smalinet](https://github.com/smalinet))
- **(handlers)** ([e6d75bd](https://github.com/babykart/gozone/commit/e6d75bd99bfeb04f23d503923df3adbd1ff4831a)) - [gozone] eliminate duplicate GetUser call in render - ([smalinet](https://github.com/smalinet))
- **(pdns)** ([9437dce](https://github.com/babykart/gozone/commit/9437dce7ca4cc84497448f383b326d0cdc1ce03e)) - [gozone] simplify URL normalization in NewClient - ([smalinet](https://github.com/smalinet))

### 📚 Documentation

- ([4eac146](https://github.com/babykart/gozone/commit/4eac1463fdc2d2b5c44de6893578e8f2daa8cab3)) - [gozone] fix diagram indentation - ([babykart](https://github.com/babykart))

### ⚡ Performance

- **(handlers)** ([c2e4556](https://github.com/babykart/gozone/commit/c2e45566ac3a100affb0cee66c025c98d7acedca)) - [gozone] parallelize independent PowerDNS calls in ViewZone - ([smalinet](https://github.com/smalinet))
- **(pdns)** ([cadcc84](https://github.com/babykart/gozone/commit/cadcc8456acc16cda2bfbc8a28192b9e2da2e868)) - [gozone] eliminate N+1 HTTP requests on zone list page - ([smalinet](https://github.com/smalinet))

<!-- generated by git-cliff -->
