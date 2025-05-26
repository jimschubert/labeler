# labeler
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fjimschubert%2Flabeler.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fjimschubert%2Flabeler?ref=badge_shield)


A labeler for GitHub issues and pull requests.

```bash
A labeler for GitHub issues and pull requests.

Usage:
  labeler [flags]

Flags:
      --config-path string   A custom config path, relative to the repository root
      --data string          A JSON string of the 'event' type (issue event or pull request event)
  -h, --help                 help for labeler
      --id int               The integer id of the issue or pull request
  -o, --owner string         GitHub Owner/Org name [GITHUB_ACTOR]
  -r, --repo string          GitHub Repo name [GITHUB_REPO]
  -t, --type string          The target event type to label (issues or pull_request) [GITHUB_EVENT_NAME]
  -v, --version              version for labeler
```

Example usage:
```bash
export GITHUB_TOKEN=yourtoken
./labeler -o jimschubert -r labeler --type pull_request --id 1
```

This will evaluate the configuration file for the repository and apply any relevant labels to PR #1.

## Configuration

The configuration file must be located in the target repository at `.github/labeler.yml` by default, and the contents must follow either the *simple* schema or the *full* schema.

The configuration file location can be modified by passing a different path to `--config-path`. This path must be relative to the repository root. All of the following would be valid possible customizations (assuming you've created a configuration file at that location):

* `.github/labeler-custom.yml`
* `tools/labeler.yml`
* `.labeler.yml`

Feel free to use one of the following schema examples to get started. 

### Simple Schema

```yaml
# labeler "simple" schema
# Comment is applied to both issues and pull requests.
# If you need a more robust solution, consider the "full" schema.
comment: |
  👍 Thanks for this!
  🏷 I have applied any labels matching special text in your issue.

  Please review the labels and make any necessary changes.

# Labels is an object where:
# - keys are labels
# - values are array of string patterns to match against title + body in issues/prs
labels:
  'bug':
    - '\bbug[s]?\b'
  'help wanted':
    - '\bhelp( wanted)?\b'
  'duplicate':
    - '\bduplicate\b'
    - '\bdupe\b'
  'enhancement':
    - '\benhancement\b'
  'question':
    - '\bquestion\b'
```

### Full Schema

```yaml
# labeler "full" schema

# enable labeler on issues, prs, or both.
enable:
  issues: true
  prs: true
# comments object allows you to specify a different message for issues and prs

comments:
  issues: |
    Thanks for opening this issue!
    I have applied any labels matching special text in your title and description.

    Please review the labels and make any necessary changes.
  prs: |
    Thanks for the contribution!
    I have applied any labels matching special text in your title and description.

    Please review the labels and make any necessary changes.

# Labels is an object where:
# - keys are labels
# - values are objects of { include: [ pattern ], exclude: [ pattern ] }
#    - pattern must be a valid regex, and is applied globally to
#      title + description of issues and/or prs (see enabled config above)
#    - 'include' patterns will associate a label if any of these patterns match
#    - 'exclude' patterns will ignore this label if any of these patterns match
labels:
  'bug':
    include:
      - '\bbug[s]?\b'
    exclude: []
  'help wanted':
    include:
      - '\bhelp( me)?\b'
    exclude:
      - '\b\[test(ing)?\]\b'
  'enhancement':
    include:
      - '\bfeat\b'
    exclude: []

```

## Build

Build a local distribution for evaluation using goreleaser.

```bash
goreleaser release --skip-publish --snapshot --rm-dist
```

This will create an executable application for your os/architecture under `dist`:

```
dist
├── labeler_darwin_amd64_v1
│   └── labeler
├── labeler_darwin_arm64
│   └── labeler
├── labeler_linux_386
│   └── labeler
├── labeler_linux_amd64_v1
│   └── labeler
├── labeler_linux_arm64
│   └── labeler
├── labeler_linux_arm_6
│   └── labeler
├── labeler_windows_amd64_v1
│   └── labeler.exe
├── labeler_windows_arm64
│   └── labeler.exe
├── labeler_windows_arm_6
│   └── labeler.exe
```

## License

The labeler project is licensed under Apache 2.0

*labeler* is a rewrite of an earlier GitHub App I wrote (see [auto-labeler](https://github.com/jimschubert/auto-labeler)). I've rewritten that app to replace the license while making the tool reusable across CI tools and operating systems.


[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fjimschubert%2Flabeler.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fjimschubert%2Flabeler?ref=badge_large)