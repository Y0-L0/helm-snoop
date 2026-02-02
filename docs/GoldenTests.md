Golden File Tests

- We use golden files to pin expected snooper output for real charts.
Golden Files live next to tests in each package under a local testdata/ directory.
Input charts live under repository-level testdata/<chart-name>.

- Update Golden Files via: go test ./... -update

JSON Structure for Results

- Path objects are serialized to a JsonPrefix ID string and a slash-aligned kinds string:
  - { "id": "/a/b/1/c", "kinds": "K/K/I/K" }
  - K = key (map field), I = index (array index), A = any (unknown/invariant)
- Results JSON:
  - { "referenced": [<PathJSON>...], "unused": [...], "undefined": [...] }
- Arrays are sorted deterministically for stable diffs.

