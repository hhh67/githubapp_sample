# githubapp_sample

GitHub Appを使用してインストールアクセストークンを発行し、以下のAPIが使用できるかの検証。

- Repositories.GetContents
- Repositories.DownloadContents
- Repositories.ListCommits
- RateLimits

## 環境変数について

key|value
-|-
GITHUB_APP_ID|GitHubAppのページから取得
GITHUB_INSTALLATION_ID|Applicationsの画面から、Configureをクリックした遷移先のURL`https://github.com/settings/installations/xxxxxxxx`のxxxxxxxxの部分
GITHUB_PRIVATE_KEY|GitHubAppのページからPrivate Keyを発行して取得