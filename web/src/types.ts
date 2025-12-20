export type CommitStats = {
  hash: string
  author: string
  date: string
  added: number
  removed: number
  message: string
  filesTouchedCount: number
}

export type FileTouchCount = {
  file: string
  count: number
}

export type Repository = {
  username: string
  repoName: string
  totalAdditions: number
  totalLines: number
  totalRemovals: number
  views: number
  linesHistogram: number[]
  totalStars: number
  totalCommits: number
}

export type GitHubRepo = {
  stargazers_count: number
  language: string
  size: number
}
