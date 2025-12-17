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
