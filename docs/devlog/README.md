# Daily Devlog Guide

This directory contains the template and instructions for generating daily devlogs for the Gemini Live Agent Challenge.

## How to use

1. Copy `DAILY_DEVLOG_TEMPLATE.md` to a new file (e.g., `2026-02-25.md`).
2. Run the commands below to collect your activity from the last 24 hours.
3. Fill in the narrative and other sections.
4. Post to dev.to!

## Activity Collection Commands

### 1. Get ISO Timestamp for 24 hours ago

**macOS (BSD date):**
```bash
SINCE=$(date -v-24H -u +"%Y-%m-%dT%H:%M:%SZ")
echo $SINCE
```

**Linux (GNU date):**
```bash
SINCE=$(date -u -d "24 hours ago" +"%Y-%m-%dT%H:%M:%SZ")
echo $SINCE
```

### 2. Git Commits (Last 24h)
```bash
git log --since="24 hours ago" --oneline --no-merges
```

### 3. GitHub PR Activity (Last 24h)
```bash
gh pr list --state all --search "updated:>$SINCE" --limit 50
```

### 4. GitHub Issue Activity (Last 24h)
```bash
gh issue list --state all --search "updated:>$SINCE" --limit 50
```

## Requirements
- Include the sentence: "I created this post for the purposes of entering the Gemini Live Agent Challenge."
- Use the hashtag: `#GeminiLiveAgentChallenge`
