import { useQuery } from '@tanstack/react-query'
import type { CommitStats, AnalyzeResponse } from '@/types'

async function analyzeRepo(username: string, repo: string) {
  const apiUrl = import.meta.env.VITE_API_URL
  const response = await fetch(`${apiUrl}/api/analyze`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ username, repo }),
  })

  if (!response.ok) {
    if (response.status === 404) {
      throw new Error('NOT_FOUND')
    }
    throw new Error('Failed to analyze repository')
  }

  const data = (await response.json()) as AnalyzeResponse

  return {
    totalAdded: data.totalAdded,
    totalRemoved: data.totalRemoved,
    totalContributors: data.totalContributors,
    totalCommits: data.totalCommits,
    commits: data.commits.map(
      (commit) =>
        ({
          hash: commit.h,
          author: commit.a,
          date: new Date(commit.d * 1000).toISOString(),
          added: commit['+'] ?? 0,
          removed: commit['-'] ?? 0,
          message: commit.m,
          filesTouchedCount: commit.f ?? 0,
        }) as CommitStats
    ),
    github: data.github,
    pullRequests: data.pullRequests || null,
  }
}

export function useRepository(
  username: string | undefined,
  repo: string | undefined
) {
  const enabled = true
  const refetchOnMount = false
  const staleTime = 5 * 60 * 1000 // 5 minutes
  const gcTime = 10 * 60 * 1000 // 10 minutes

  const query = useQuery({
    queryKey: ['repo', username, repo],
    queryFn: () => analyzeRepo(username!, repo!),
    enabled: enabled && !!username && !!repo,
    retry: 3,
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
    refetchOnMount,
    staleTime,
    gcTime,
  })

  return {
    ...query,
    data: query.data,
    isNotFound:
      query.error instanceof Error && query.error.message === 'NOT_FOUND',
  }
}
