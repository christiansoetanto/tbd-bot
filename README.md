# TBD-BOT
Because I couldn't think of a name at time of writing the code.

## Features
- Vetting system
  - Verification and role assigning (/sdverify)
  - Vetting response's secret code detector (the INRI code)
  - Welcome message
- Daily Catholic
  - Office of Readings
  - Liturgical celebration for today
  - Friday abstinence memes
- Q&A
  - #religious-questions, #religious-discussions, #religious-discussions-2, and #answered-questions system where, as the name implies, answered questions will be moved to #answered-questions so that unanswered questions can have more exposure.
## Monitoring & Grafana Access
Grafana is exposed locally on port `3000` (bound to `127.0.0.1:3000` for security):
- Local URL: `http://127.0.0.1:3000` (accessible only from the Mac Mini itself; Grafana is bound to localhost for security)
- Credentials: Default `admin` / `admin` (change upon first login or set via `.env`)
- Pre-configured Grafana dashboard auto-loaded from `grafana/provisioning/dashboards/bot-dashboard.json`.

## Self-Hosted Runner Security
To protect the local host environment and self-hosted runner from arbitrary code execution via external Fork Pull Requests:
1. Open your repository on GitHub.
2. Navigate to **Settings** -> **Actions** -> **General**.
3. Under **Fork pull request workflows from outside collaborators**, choose **Require approval for all outside collaborators** (or disable fork PR execution on self-hosted runners).
4. Ensure workflow jobs targeting `runs-on: self-hosted` are tied strictly to the `main` branch.

## [![Repography logo](https://images.repography.com/logo.svg)](https://repography.com) / Recent activity [![Time period](https://images.repography.com/26965455/christiansoetanto/tbd-bot/recent-activity/XSKu8NuvUO-wMkvxmAk0-Sh04dEgjpwow1r37BcSWVk/ZHy8hl1p33C5CZb9geQ1DTotvDo6YiFPQ9owxkhS1qU_badge.svg)](https://repography.com)
[![Timeline graph](https://images.repography.com/26965455/christiansoetanto/tbd-bot/recent-activity/XSKu8NuvUO-wMkvxmAk0-Sh04dEgjpwow1r37BcSWVk/ZHy8hl1p33C5CZb9geQ1DTotvDo6YiFPQ9owxkhS1qU_timeline.svg)](https://github.com/christiansoetanto/tbd-bot/commits)
[![Issue status graph](https://images.repography.com/26965455/christiansoetanto/tbd-bot/recent-activity/XSKu8NuvUO-wMkvxmAk0-Sh04dEgjpwow1r37BcSWVk/ZHy8hl1p33C5CZb9geQ1DTotvDo6YiFPQ9owxkhS1qU_issues.svg)](https://github.com/christiansoetanto/tbd-bot/issues)
[![Pull request status graph](https://images.repography.com/26965455/christiansoetanto/tbd-bot/recent-activity/XSKu8NuvUO-wMkvxmAk0-Sh04dEgjpwow1r37BcSWVk/ZHy8hl1p33C5CZb9geQ1DTotvDo6YiFPQ9owxkhS1qU_prs.svg)](https://github.com/christiansoetanto/tbd-bot/pulls)
[![Trending topics](https://images.repography.com/26965455/christiansoetanto/tbd-bot/recent-activity/XSKu8NuvUO-wMkvxmAk0-Sh04dEgjpwow1r37BcSWVk/ZHy8hl1p33C5CZb9geQ1DTotvDo6YiFPQ9owxkhS1qU_words.svg)](https://github.com/christiansoetanto/tbd-bot/commits)
[![Top contributors](https://images.repography.com/26965455/christiansoetanto/tbd-bot/recent-activity/XSKu8NuvUO-wMkvxmAk0-Sh04dEgjpwow1r37BcSWVk/ZHy8hl1p33C5CZb9geQ1DTotvDo6YiFPQ9owxkhS1qU_users.svg)](https://github.com/christiansoetanto/tbd-bot/graphs/contributors)
[![Activity map](https://images.repography.com/26965455/christiansoetanto/tbd-bot/recent-activity/XSKu8NuvUO-wMkvxmAk0-Sh04dEgjpwow1r37BcSWVk/ZHy8hl1p33C5CZb9geQ1DTotvDo6YiFPQ9owxkhS1qU_map.svg)](https://github.com/christiansoetanto/tbd-bot/commits)



## [![Repography logo](https://images.repography.com/logo.svg)](https://repography.com) / Structure
[![Structure](https://images.repography.com/26965455/christiansoetanto/tbd-bot/structure/XSKu8NuvUO-wMkvxmAk0-Sh04dEgjpwow1r37BcSWVk/FNR-8bU2tOwB9W1WX1pp7RKuhSjJagCinaWXKcfjiXk_table.svg)](https://github.com/christiansoetanto/tbd-bot)

