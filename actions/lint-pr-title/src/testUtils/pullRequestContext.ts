import { GitHubPayload, newContext } from ".";

import { Context } from "@actions/github/lib/context";
import { PullRequestOpenedEvent } from "@octokit/webhooks-types";

const payload: GitHubPayload<PullRequestOpenedEvent> = {
  token: "ghp_tokengoeshere",
  job: "pull-request",
  ref: "refs/heads/main",
  sha: "0123456789abcdef0123456789abcdef01234567",
  repository: "someorg/somerepo",
  repository_owner: "someorg",
  repository_owner_id: "9876543",
  repositoryUrl: "git://github.com/someorg/somerepo.git",
  run_id: "11111111111",
  run_number: "1",
  retention_days: "90",
  run_attempt: "1",
  artifact_cache_size_limit: "10",
  repository_visibility: "private",
  "repo-self-hosted-runners-disabled": false,
  "enterprise-managed-business-id": "",
  repository_id: "111222333",
  actor_id: "123456",
  actor: "someuser",
  triggering_actor: "someuser",
  workflow: ".github/workflows/pull-request.yml",
  head_ref: "someuser/branch",
  base_ref: "main",
  event_name: "pull_request_target",
  event: {
    action: "opened",
    enterprise: {
      avatar_url: "https://avatars.githubusercontent.com/b/123456?v=4",
      created_at: "2024-02-29T23:01:47Z",
      description: "someent is the best enterprise ever.",
      html_url: "https://github.com/enterprises/someent",
      id: 123456,
      name: "someent enterprise",
      node_id: "blah",
      slug: "someorg",
      updated_at: "2024-04-25T13:56:52Z",
      website_url: "https://someorg.com",
    },
    number: 43,
    organization: {
      avatar_url: "https://avatars.githubusercontent.com/u/9876543?v=4",
      description: "someorg is the best org.",
      events_url: "https://api.github.com/orgs/someorg/events",
      hooks_url: "https://api.github.com/orgs/someorg/hooks",
      id: 9876543,
      issues_url: "https://api.github.com/orgs/someorg/issues",
      login: "someorg",
      members_url: "https://api.github.com/orgs/someorg/members{/member}",
      node_id: "bleh=",
      public_members_url:
        "https://api.github.com/orgs/someorg/public_members{/member}",
      repos_url: "https://api.github.com/orgs/someorg/repos",
      url: "https://api.github.com/orgs/someorg",
    },
    pull_request: {
      _links: {
        comments: {
          href: "https://api.github.com/repos/someorg/somerepo/issues/43/comments",
        },
        commits: {
          href: "https://api.github.com/repos/someorg/somerepo/pulls/43/commits",
        },
        html: {
          href: "https://github.com/someorg/somerepo/pull/43",
        },
        issue: {
          href: "https://api.github.com/repos/someorg/somerepo/issues/43",
        },
        review_comment: {
          href: "https://api.github.com/repos/someorg/somerepo/pulls/comments{/number}",
        },
        review_comments: {
          href: "https://api.github.com/repos/someorg/somerepo/pulls/43/comments",
        },
        self: {
          href: "https://api.github.com/repos/someorg/somerepo/pulls/43",
        },
        statuses: {
          href: "https://api.github.com/repos/someorg/somerepo/statuses/e41343a54606009d64b2236003a3d4ea5e7a7ee8",
        },
      },
      active_lock_reason: null,
      additions: 0,
      assignee: null,
      assignees: [],
      author_association: "CONTRIBUTOR",
      auto_merge: null,
      base: {
        label: "someorg:main",
        ref: "main",
        repo: {
          allow_auto_merge: false,
          allow_forking: false,
          allow_merge_commit: true,
          allow_rebase_merge: true,
          allow_squash_merge: true,
          allow_update_branch: false,
          archive_url:
            "https://api.github.com/repos/someorg/somerepo/{archive_format}{/ref}",
          archived: false,
          assignees_url:
            "https://api.github.com/repos/someorg/somerepo/assignees{/user}",
          blobs_url:
            "https://api.github.com/repos/someorg/somerepo/git/blobs{/sha}",
          branches_url:
            "https://api.github.com/repos/someorg/somerepo/branches{/branch}",
          clone_url: "https://github.com/someorg/somerepo.git",
          collaborators_url:
            "https://api.github.com/repos/someorg/somerepo/collaborators{/collaborator}",
          comments_url:
            "https://api.github.com/repos/someorg/somerepo/comments{/number}",
          commits_url:
            "https://api.github.com/repos/someorg/somerepo/commits{/sha}",
          compare_url:
            "https://api.github.com/repos/someorg/somerepo/compare/{base}...{head}",
          contents_url:
            "https://api.github.com/repos/someorg/somerepo/contents/{+path}",
          contributors_url:
            "https://api.github.com/repos/someorg/somerepo/contributors",
          created_at: "2024-01-26T13:35:52Z",
          custom_properties: {},
          default_branch: "main",
          delete_branch_on_merge: false,
          deployments_url:
            "https://api.github.com/repos/someorg/somerepo/deployments",
          description: null,
          disabled: false,
          downloads_url:
            "https://api.github.com/repos/someorg/somerepo/downloads",
          events_url: "https://api.github.com/repos/someorg/somerepo/events",
          fork: false,
          forks: 0,
          forks_count: 0,
          forks_url: "https://api.github.com/repos/someorg/somerepo/forks",
          full_name: "someorg/somerepo",
          git_commits_url:
            "https://api.github.com/repos/someorg/somerepo/git/commits{/sha}",
          git_refs_url:
            "https://api.github.com/repos/someorg/somerepo/git/refs{/sha}",
          git_tags_url:
            "https://api.github.com/repos/someorg/somerepo/git/tags{/sha}",
          git_url: "git://github.com/someorg/somerepo.git",
          has_discussions: false,
          has_downloads: true,
          has_issues: true,
          has_pages: false,
          has_projects: true,
          has_wiki: true,
          homepage: null,
          hooks_url: "https://api.github.com/repos/someorg/somerepo/hooks",
          html_url: "https://github.com/someorg/somerepo",
          id: 111222333,
          is_template: false,
          issue_comment_url:
            "https://api.github.com/repos/someorg/somerepo/issues/comments{/number}",
          issue_events_url:
            "https://api.github.com/repos/someorg/somerepo/issues/events{/number}",
          issues_url:
            "https://api.github.com/repos/someorg/somerepo/issues{/number}",
          keys_url:
            "https://api.github.com/repos/someorg/somerepo/keys{/key_id}",
          labels_url:
            "https://api.github.com/repos/someorg/somerepo/labels{/name}",
          language: null,
          languages_url:
            "https://api.github.com/repos/someorg/somerepo/languages",
          license: null,
          merge_commit_message: "PR_TITLE",
          merge_commit_title: "MERGE_MESSAGE",
          merges_url: "https://api.github.com/repos/someorg/somerepo/merges",
          milestones_url:
            "https://api.github.com/repos/someorg/somerepo/milestones{/number}",
          mirror_url: null,
          name: "somerepo",
          node_id: "wibble",
          notifications_url:
            "https://api.github.com/repos/someorg/somerepo/notifications{?since,all,participating}",
          open_issues: 4,
          open_issues_count: 4,
          owner: {
            avatar_url: "https://avatars.githubusercontent.com/u/9876543?v=4",
            events_url: "https://api.github.com/users/someorg/events{/privacy}",
            followers_url: "https://api.github.com/users/someorg/followers",
            following_url:
              "https://api.github.com/users/someorg/following{/other_user}",
            gists_url: "https://api.github.com/users/someorg/gists{/gist_id}",
            gravatar_id: "",
            html_url: "https://github.com/someorg",
            id: 9876543,
            login: "someorg",
            node_id: "=",
            organizations_url: "https://api.github.com/users/someorg/orgs",
            received_events_url:
              "https://api.github.com/users/someorg/received_events",
            repos_url: "https://api.github.com/users/someorg/repos",
            site_admin: false,
            starred_url:
              "https://api.github.com/users/someorg/starred{/owner}{/repo}",
            subscriptions_url:
              "https://api.github.com/users/someorg/subscriptions",
            type: "Organization",
            url: "https://api.github.com/users/someorg",
          },
          private: true,
          pulls_url:
            "https://api.github.com/repos/someorg/somerepo/pulls{/number}",
          pushed_at: "2024-09-11T13:38:19Z",
          releases_url:
            "https://api.github.com/repos/someorg/somerepo/releases{/id}",
          size: 109,
          squash_merge_commit_message: "COMMIT_MESSAGES",
          squash_merge_commit_title: "COMMIT_OR_PR_TITLE",
          ssh_url: "git@github.com:someorg/somerepo.git",
          stargazers_count: 0,
          stargazers_url:
            "https://api.github.com/repos/someorg/somerepo/stargazers",
          statuses_url:
            "https://api.github.com/repos/someorg/somerepo/statuses/{sha}",
          subscribers_url:
            "https://api.github.com/repos/someorg/somerepo/subscribers",
          subscription_url:
            "https://api.github.com/repos/someorg/somerepo/subscription",
          svn_url: "https://github.com/someorg/somerepo",
          tags_url: "https://api.github.com/repos/someorg/somerepo/tags",
          teams_url: "https://api.github.com/repos/someorg/somerepo/teams",
          topics: [],
          trees_url:
            "https://api.github.com/repos/someorg/somerepo/git/trees{/sha}",
          updated_at: "2024-09-09T10:45:26Z",
          url: "https://api.github.com/repos/someorg/somerepo",
          use_squash_pr_title_as_default: false,
          visibility: "private",
          watchers: 0,
          watchers_count: 0,
          web_commit_signoff_required: false,
        },
        sha: "0123456789abcdef0123456789abcdef01234567",
        user: {
          avatar_url: "https://avatars.githubusercontent.com/u/9876543?v=4",
          events_url: "https://api.github.com/users/someorg/events{/privacy}",
          followers_url: "https://api.github.com/users/someorg/followers",
          following_url:
            "https://api.github.com/users/someorg/following{/other_user}",
          gists_url: "https://api.github.com/users/someorg/gists{/gist_id}",
          gravatar_id: "",
          html_url: "https://github.com/someorg",
          id: 9876543,
          login: "someorg",
          node_id: "zomg",
          organizations_url: "https://api.github.com/users/someorg/orgs",
          received_events_url:
            "https://api.github.com/users/someorg/received_events",
          repos_url: "https://api.github.com/users/someorg/repos",
          site_admin: false,
          starred_url:
            "https://api.github.com/users/someorg/starred{/owner}{/repo}",
          subscriptions_url:
            "https://api.github.com/users/someorg/subscriptions",
          type: "Organization",
          url: "https://api.github.com/users/someorg",
        },
      },
      body: null,
      changed_files: 0,
      closed_at: null,
      comments: 0,
      comments_url:
        "https://api.github.com/repos/someorg/somerepo/issues/43/comments",
      commits: 1,
      commits_url:
        "https://api.github.com/repos/someorg/somerepo/pulls/43/commits",
      created_at: "2024-09-11T13:38:20Z",
      deletions: 0,
      diff_url: "https://github.com/someorg/somerepo/pull/43.diff",
      draft: false,
      head: {
        label: "someorg:someuser/branch",
        ref: "someuser/branch",
        repo: {
          allow_auto_merge: false,
          allow_forking: false,
          allow_merge_commit: true,
          allow_rebase_merge: true,
          allow_squash_merge: true,
          allow_update_branch: false,
          archive_url:
            "https://api.github.com/repos/someorg/somerepo/{archive_format}{/ref}",
          archived: false,
          assignees_url:
            "https://api.github.com/repos/someorg/somerepo/assignees{/user}",
          blobs_url:
            "https://api.github.com/repos/someorg/somerepo/git/blobs{/sha}",
          branches_url:
            "https://api.github.com/repos/someorg/somerepo/branches{/branch}",
          clone_url: "https://github.com/someorg/somerepo.git",
          collaborators_url:
            "https://api.github.com/repos/someorg/somerepo/collaborators{/collaborator}",
          comments_url:
            "https://api.github.com/repos/someorg/somerepo/comments{/number}",
          commits_url:
            "https://api.github.com/repos/someorg/somerepo/commits{/sha}",
          compare_url:
            "https://api.github.com/repos/someorg/somerepo/compare/{base}...{head}",
          contents_url:
            "https://api.github.com/repos/someorg/somerepo/contents/{+path}",
          contributors_url:
            "https://api.github.com/repos/someorg/somerepo/contributors",
          created_at: "2024-01-26T13:35:52Z",
          default_branch: "main",
          delete_branch_on_merge: false,
          deployments_url:
            "https://api.github.com/repos/someorg/somerepo/deployments",
          description: null,
          disabled: false,
          downloads_url:
            "https://api.github.com/repos/someorg/somerepo/downloads",
          events_url: "https://api.github.com/repos/someorg/somerepo/events",
          fork: false,
          forks: 0,
          forks_count: 0,
          forks_url: "https://api.github.com/repos/someorg/somerepo/forks",
          full_name: "someorg/somerepo",
          git_commits_url:
            "https://api.github.com/repos/someorg/somerepo/git/commits{/sha}",
          git_refs_url:
            "https://api.github.com/repos/someorg/somerepo/git/refs{/sha}",
          git_tags_url:
            "https://api.github.com/repos/someorg/somerepo/git/tags{/sha}",
          git_url: "git://github.com/someorg/somerepo.git",
          has_discussions: false,
          has_downloads: true,
          has_issues: true,
          has_pages: false,
          has_projects: true,
          has_wiki: true,
          homepage: null,
          hooks_url: "https://api.github.com/repos/someorg/somerepo/hooks",
          html_url: "https://github.com/someorg/somerepo",
          id: 111222333,
          is_template: false,
          issue_comment_url:
            "https://api.github.com/repos/someorg/somerepo/issues/comments{/number}",
          issue_events_url:
            "https://api.github.com/repos/someorg/somerepo/issues/events{/number}",
          issues_url:
            "https://api.github.com/repos/someorg/somerepo/issues{/number}",
          keys_url:
            "https://api.github.com/repos/someorg/somerepo/keys{/key_id}",
          labels_url:
            "https://api.github.com/repos/someorg/somerepo/labels{/name}",
          language: null,
          languages_url:
            "https://api.github.com/repos/someorg/somerepo/languages",
          license: null,
          merge_commit_message: "PR_TITLE",
          merge_commit_title: "MERGE_MESSAGE",
          merges_url: "https://api.github.com/repos/someorg/somerepo/merges",
          milestones_url:
            "https://api.github.com/repos/someorg/somerepo/milestones{/number}",
          mirror_url: null,
          name: "somerepo",
          node_id: "yikes",
          notifications_url:
            "https://api.github.com/repos/someorg/somerepo/notifications{?since,all,participating}",
          open_issues: 4,
          open_issues_count: 4,
          owner: {
            avatar_url: "https://avatars.githubusercontent.com/u/9876543?v=4",
            events_url: "https://api.github.com/users/someorg/events{/privacy}",
            followers_url: "https://api.github.com/users/someorg/followers",
            following_url:
              "https://api.github.com/users/someorg/following{/other_user}",
            gists_url: "https://api.github.com/users/someorg/gists{/gist_id}",
            gravatar_id: "",
            html_url: "https://github.com/someorg",
            id: 9876543,
            login: "someorg",
            node_id: "yowza=",
            organizations_url: "https://api.github.com/users/someorg/orgs",
            received_events_url:
              "https://api.github.com/users/someorg/received_events",
            repos_url: "https://api.github.com/users/someorg/repos",
            site_admin: false,
            starred_url:
              "https://api.github.com/users/someorg/starred{/owner}{/repo}",
            subscriptions_url:
              "https://api.github.com/users/someorg/subscriptions",
            type: "Organization",
            url: "https://api.github.com/users/someorg",
          },
          private: true,
          pulls_url:
            "https://api.github.com/repos/someorg/somerepo/pulls{/number}",
          pushed_at: "2024-09-11T13:38:19Z",
          releases_url:
            "https://api.github.com/repos/someorg/somerepo/releases{/id}",
          size: 109,
          squash_merge_commit_message: "COMMIT_MESSAGES",
          squash_merge_commit_title: "COMMIT_OR_PR_TITLE",
          ssh_url: "git@github.com:someorg/somerepo.git",
          stargazers_count: 0,
          stargazers_url:
            "https://api.github.com/repos/someorg/somerepo/stargazers",
          statuses_url:
            "https://api.github.com/repos/someorg/somerepo/statuses/{sha}",
          subscribers_url:
            "https://api.github.com/repos/someorg/somerepo/subscribers",
          subscription_url:
            "https://api.github.com/repos/someorg/somerepo/subscription",
          svn_url: "https://github.com/someorg/somerepo",
          tags_url: "https://api.github.com/repos/someorg/somerepo/tags",
          teams_url: "https://api.github.com/repos/someorg/somerepo/teams",
          topics: [],
          trees_url:
            "https://api.github.com/repos/someorg/somerepo/git/trees{/sha}",
          updated_at: "2024-09-09T10:45:26Z",
          url: "https://api.github.com/repos/someorg/somerepo",
          use_squash_pr_title_as_default: false,
          visibility: "private",
          watchers: 0,
          watchers_count: 0,
          web_commit_signoff_required: false,
        },
        sha: "e41343a54606009d64b2236003a3d4ea5e7a7ee8",
        user: {
          avatar_url: "https://avatars.githubusercontent.com/u/9876543?v=4",
          events_url: "https://api.github.com/users/someorg/events{/privacy}",
          followers_url: "https://api.github.com/users/someorg/followers",
          following_url:
            "https://api.github.com/users/someorg/following{/other_user}",
          gists_url: "https://api.github.com/users/someorg/gists{/gist_id}",
          gravatar_id: "",
          html_url: "https://github.com/someorg",
          id: 9876543,
          login: "someorg",
          node_id: "help=",
          organizations_url: "https://api.github.com/users/someorg/orgs",
          received_events_url:
            "https://api.github.com/users/someorg/received_events",
          repos_url: "https://api.github.com/users/someorg/repos",
          site_admin: false,
          starred_url:
            "https://api.github.com/users/someorg/starred{/owner}{/repo}",
          subscriptions_url:
            "https://api.github.com/users/someorg/subscriptions",
          type: "Organization",
          url: "https://api.github.com/users/someorg",
        },
      },
      html_url: "https://github.com/someorg/somerepo/pull/43",
      id: 1231231231,
      issue_url: "https://api.github.com/repos/someorg/somerepo/issues/43",
      labels: [],
      locked: false,
      maintainer_can_modify: false,
      merge_commit_sha: "9999999999999999999999999999999999999999",
      mergeable: true,
      mergeable_state: "unstable",
      merged: false,
      merged_at: null,
      merged_by: null,
      milestone: null,
      node_id: "argh",
      number: 43,
      patch_url: "https://github.com/someorg/somerepo/pull/43.patch",
      rebaseable: true,
      requested_reviewers: [],
      requested_teams: [],
      review_comment_url:
        "https://api.github.com/repos/someorg/somerepo/pulls/comments{/number}",
      review_comments: 0,
      review_comments_url:
        "https://api.github.com/repos/someorg/somerepo/pulls/43/comments",
      state: "open",
      statuses_url:
        "https://api.github.com/repos/someorg/somerepo/statuses/9999999999999999999999999999999999999999",
      title: "empty",
      updated_at: "2024-09-11T13:38:20Z",
      url: "https://api.github.com/repos/someorg/somerepo/pulls/43",
      user: {
        avatar_url: "https://avatars.githubusercontent.com/u/123456?v=4",
        events_url: "https://api.github.com/users/someuser/events{/privacy}",
        followers_url: "https://api.github.com/users/someuser/followers",
        following_url:
          "https://api.github.com/users/someuser/following{/other_user}",
        gists_url: "https://api.github.com/users/someuser/gists{/gist_id}",
        gravatar_id: "",
        html_url: "https://github.com/someuser",
        id: 123456,
        login: "someuser",
        node_id: "f00==",
        organizations_url: "https://api.github.com/users/someuser/orgs",
        received_events_url:
          "https://api.github.com/users/someuser/received_events",
        repos_url: "https://api.github.com/users/someuser/repos",
        site_admin: false,
        starred_url:
          "https://api.github.com/users/someuser/starred{/owner}{/repo}",
        subscriptions_url:
          "https://api.github.com/users/someuser/subscriptions",
        type: "User",
        url: "https://api.github.com/users/someuser",
      },
    },
    repository: {
      allow_forking: false,
      archive_url:
        "https://api.github.com/repos/someorg/somerepo/{archive_format}{/ref}",
      archived: false,
      assignees_url:
        "https://api.github.com/repos/someorg/somerepo/assignees{/user}",
      blobs_url:
        "https://api.github.com/repos/someorg/somerepo/git/blobs{/sha}",
      branches_url:
        "https://api.github.com/repos/someorg/somerepo/branches{/branch}",
      clone_url: "https://github.com/someorg/somerepo.git",
      collaborators_url:
        "https://api.github.com/repos/someorg/somerepo/collaborators{/collaborator}",
      comments_url:
        "https://api.github.com/repos/someorg/somerepo/comments{/number}",
      commits_url:
        "https://api.github.com/repos/someorg/somerepo/commits{/sha}",
      compare_url:
        "https://api.github.com/repos/someorg/somerepo/compare/{base}...{head}",
      contents_url:
        "https://api.github.com/repos/someorg/somerepo/contents/{+path}",
      contributors_url:
        "https://api.github.com/repos/someorg/somerepo/contributors",
      created_at: "2024-01-26T13:35:52Z",
      custom_properties: {},
      default_branch: "main",
      deployments_url:
        "https://api.github.com/repos/someorg/somerepo/deployments",
      description: null,
      disabled: false,
      downloads_url: "https://api.github.com/repos/someorg/somerepo/downloads",
      events_url: "https://api.github.com/repos/someorg/somerepo/events",
      fork: false,
      forks: 0,
      forks_count: 0,
      forks_url: "https://api.github.com/repos/someorg/somerepo/forks",
      full_name: "someorg/somerepo",
      git_commits_url:
        "https://api.github.com/repos/someorg/somerepo/git/commits{/sha}",
      git_refs_url:
        "https://api.github.com/repos/someorg/somerepo/git/refs{/sha}",
      git_tags_url:
        "https://api.github.com/repos/someorg/somerepo/git/tags{/sha}",
      git_url: "git://github.com/someorg/somerepo.git",
      has_discussions: false,
      has_downloads: true,
      has_issues: true,
      has_pages: false,
      has_projects: true,
      has_wiki: true,
      homepage: null,
      hooks_url: "https://api.github.com/repos/someorg/somerepo/hooks",
      html_url: "https://github.com/someorg/somerepo",
      id: 111222333,
      is_template: false,
      issue_comment_url:
        "https://api.github.com/repos/someorg/somerepo/issues/comments{/number}",
      issue_events_url:
        "https://api.github.com/repos/someorg/somerepo/issues/events{/number}",
      issues_url:
        "https://api.github.com/repos/someorg/somerepo/issues{/number}",
      keys_url: "https://api.github.com/repos/someorg/somerepo/keys{/key_id}",
      labels_url: "https://api.github.com/repos/someorg/somerepo/labels{/name}",
      language: null,
      languages_url: "https://api.github.com/repos/someorg/somerepo/languages",
      license: null,
      merges_url: "https://api.github.com/repos/someorg/somerepo/merges",
      milestones_url:
        "https://api.github.com/repos/someorg/somerepo/milestones{/number}",
      mirror_url: null,
      name: "somerepo",
      node_id: "R_kgDOLJ-ebA",
      notifications_url:
        "https://api.github.com/repos/someorg/somerepo/notifications{?since,all,participating}",
      open_issues: 4,
      open_issues_count: 4,
      owner: {
        avatar_url: "https://avatars.githubusercontent.com/u/9876543?v=4",
        events_url: "https://api.github.com/users/someorg/events{/privacy}",
        followers_url: "https://api.github.com/users/someorg/followers",
        following_url:
          "https://api.github.com/users/someorg/following{/other_user}",
        gists_url: "https://api.github.com/users/someorg/gists{/gist_id}",
        gravatar_id: "",
        html_url: "https://github.com/someorg",
        id: 9876543,
        login: "someorg",
        node_id: "howmanyaretheretoreplace=",
        organizations_url: "https://api.github.com/users/someorg/orgs",
        received_events_url:
          "https://api.github.com/users/someorg/received_events",
        repos_url: "https://api.github.com/users/someorg/repos",
        site_admin: false,
        starred_url:
          "https://api.github.com/users/someorg/starred{/owner}{/repo}",
        subscriptions_url: "https://api.github.com/users/someorg/subscriptions",
        type: "Organization",
        url: "https://api.github.com/users/someorg",
      },
      private: true,
      pulls_url: "https://api.github.com/repos/someorg/somerepo/pulls{/number}",
      pushed_at: "2024-09-11T13:38:19Z",
      releases_url:
        "https://api.github.com/repos/someorg/somerepo/releases{/id}",
      size: 109,
      ssh_url: "git@github.com:someorg/somerepo.git",
      stargazers_count: 0,
      stargazers_url:
        "https://api.github.com/repos/someorg/somerepo/stargazers",
      statuses_url:
        "https://api.github.com/repos/someorg/somerepo/statuses/{sha}",
      subscribers_url:
        "https://api.github.com/repos/someorg/somerepo/subscribers",
      subscription_url:
        "https://api.github.com/repos/someorg/somerepo/subscription",
      svn_url: "https://github.com/someorg/somerepo",
      tags_url: "https://api.github.com/repos/someorg/somerepo/tags",
      teams_url: "https://api.github.com/repos/someorg/somerepo/teams",
      topics: [],
      trees_url:
        "https://api.github.com/repos/someorg/somerepo/git/trees{/sha}",
      updated_at: "2024-09-09T10:45:26Z",
      url: "https://api.github.com/repos/someorg/somerepo",
      visibility: "private",
      watchers: 0,
      watchers_count: 0,
      web_commit_signoff_required: false,
    },
    sender: {
      avatar_url: "https://avatars.githubusercontent.com/u/123456?v=4",
      events_url: "https://api.github.com/users/someuser/events{/privacy}",
      followers_url: "https://api.github.com/users/someuser/followers",
      following_url:
        "https://api.github.com/users/someuser/following{/other_user}",
      gists_url: "https://api.github.com/users/someuser/gists{/gist_id}",
      gravatar_id: "",
      html_url: "https://github.com/someuser",
      id: 123456,
      login: "someuser",
      node_id: "f00==",
      organizations_url: "https://api.github.com/users/someuser/orgs",
      received_events_url:
        "https://api.github.com/users/someuser/received_events",
      repos_url: "https://api.github.com/users/someuser/repos",
      site_admin: false,
      starred_url:
        "https://api.github.com/users/someuser/starred{/owner}{/repo}",
      subscriptions_url: "https://api.github.com/users/someuser/subscriptions",
      type: "User",
      url: "https://api.github.com/users/someuser",
    },
  },
  server_url: "https://github.com",
  api_url: "https://api.github.com",
  graphql_url: "https://api.github.com/graphql",
  ref_name: "main",
  ref_protected: true,
  ref_type: "branch",
  secret_source: "Actions",
  workflow_ref:
    "someorg/somerepo/.github/workflows/pull-request.yml@refs/heads/main",
  workflow_sha: "0123456789abcdef0123456789abcdef01234567",
  workspace: "/opt/actions-runner/_work/somerepo/somerepo",
  event_path: "/opt/actions-runner/_work/_temp/_github_workflow/event.json",
  path: "/opt/actions-runner/_work/_temp/_runner_file_commands/add_path_65bcb106-8b8e-4275-8061-2080df62c11e",
  env: "/opt/actions-runner/_work/_temp/_runner_file_commands/set_env_65bcb106-8b8e-4275-8061-2080df62c11e",
  step_summary:
    "/opt/actions-runner/_work/_temp/_runner_file_commands/step_summary_65bcb106-8b8e-4275-8061-2080df62c11e",
  state:
    "/opt/actions-runner/_work/_temp/_runner_file_commands/save_state_65bcb106-8b8e-4275-8061-2080df62c11e",
  output:
    "/opt/actions-runner/_work/_temp/_runner_file_commands/set_output_65bcb106-8b8e-4275-8061-2080df62c11e",
  action: "__run_2",
  action_repository: "",
  action_ref: "",
} as const;

export function newContextFromPullRequest(
  title?: string,
  body?: string,
): Context {
  const { sha, ref } = payload.event.pull_request.head;

  const ctx = newContext(payload, sha, ref);
  if (title) {
    (ctx.payload as typeof payload.event).pull_request.title = title;
  }
  if (body) {
    (ctx.payload as typeof payload.event).pull_request.body = body;
  }

  return ctx;
}
