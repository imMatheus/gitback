import { Link, useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { CommitGraph } from '../components/commit-graph'
import type { CommitStats } from '@/types'
import { LoadingAnimation } from '@/components/loading-animation'

async function analyzeRepo(username: string, repo: string) {
  const response = await fetch('http://localhost:8080/api/analyze', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ username, repo }),
  })

  if (!response.ok) {
    throw new Error('Failed to analyze repository')
  }

  return response.json() as Promise<{
    message: string
    totalAdded: number
    totalRemoved: number
    stats: CommitStats[]
  }>
}

export default function Repo() {
  const { username, repo } = useParams<{ username: string; repo: string }>()

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ['analyze', username, repo],
    queryFn: () => analyzeRepo(username!, repo!),
    enabled: !!username && !!repo,
  })

  if (isLoading) {
    return (
      <div className="flex h-screen w-full items-center justify-center">
        <LoadingAnimation />
      </div>
    )
  }

  if (isError || !data) {
    return (
      <div>
        Error: {error instanceof Error ? error.message : 'Failed to analyze'}
      </div>
    )
  }

  return (
    <div className="min-h-screen py-4">
      <div className="mx-auto max-w-6xl px-4">
        <div className="mb-4 flex items-center">
          <Link to={'/'} className="text-sm hover:underline">
            ← Search another one
          </Link>
        </div>
        <div className="mb-3">
          <h1 className="mb-3 text-5xl font-black tracking-tight">{repo}</h1>
          <h3 className="text-2xl font-medium tracking-tight">{username}</h3>
        </div>

        <CommitGraph
          stats={data.stats}
          totalAdded={data.totalAdded}
          totalRemoved={data.totalRemoved}
        />
      </div>

      <div className="mx-auto grid max-w-6xl grid-cols-2 px-4">
        <div className="border-obsidian-field border-r-2 p-10">
          <h3 className="text-3xl font-bold">Craziest week</h3>
          <p className="text-ion-drift mb-4 text-lg font-medium tracking-wide">
            The week with the most commits
          </p>

          <h3 className="mb-5 text-4xl font-black">5,432 commits</h3>
          <div className="grid grid-cols-[1fr_auto] items-center gap-x-4 gap-y-2">
            <div className="bg-core-flux flex h-12 items-center px-4">
              <p className="text-obsidian-field text-lg font-bold">Monday</p>
            </div>
            <p className="text-lg font-bold">823 commits</p>

            <div
              className="bg-core-flux flex h-12 items-center px-4"
              style={{ width: '80%' }}
            >
              <p className="text-obsidian-field text-lg font-bold">Monday</p>
            </div>
            <p className="text-lg font-bold">823 commits</p>

            <div
              className="bg-core-flux flex h-12 items-center px-4"
              style={{ width: '60%' }}
            >
              <p className="text-obsidian-field text-lg font-bold">Monday</p>
            </div>
            <p className="text-lg font-bold">823 commits</p>
          </div>
        </div>
      </div>
    </div>
  )
}
