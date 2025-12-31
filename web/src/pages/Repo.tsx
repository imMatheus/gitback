import { Link, useParams } from 'react-router'
import { CommitGraph } from '../components/commit-graph'
import { LoadingAnimation } from '@/components/loading-animation'
import NotFound from './NotFound'
import { TopContributors } from '@/components/top-contributors'
import { CommitWordCloud } from '@/components/commit-wordcloud'
import { FileCountDistribution } from '@/components/file-count-distributionProps'
import { CommitGrid } from '@/components/commit-grid'
import { BiggestCommits } from '@/components/biggest-commits'
import { TopGitHubPRs } from '@/components/top-github-prs'
import { OverviewRecap } from '@/components/overview-recap'
import { CraziestWeek } from '@/components/CraziestWeek'
import { useRepository } from '@/hooks/useRepository'

const THIS_YEAR = 2025

export default function Repo() {
  const { username, repo } = useParams<{ username: string; repo: string }>()

  const { data, isLoading, isError, isNotFound } = useRepository(username, repo)

  if (isLoading) {
    return (
      <div className="flex h-screen w-full items-center justify-center">
        <LoadingAnimation />
      </div>
    )
  }

  if (!username || !repo) {
    return <NotFound isRepo={true} />
  }

  if (isNotFound) {
    return <NotFound isRepo={true} />
  }

  if (isError || !data) {
    return (
      <div className="flex h-screen w-full items-center justify-center">
        <div className="text-center">
          <h2 className="mb-4 text-2xl font-bold text-red-600">
            Analysis Failed
          </h2>
          <p className="mb-4 text-gray-600">
            Failed to analyze repository. Please try again later.
          </p>
          <Link
            to="/"
            className="rounded bg-blue-500 px-4 py-2 text-white hover:bg-blue-600"
          >
            Go Home
          </Link>
        </div>
      </div>
    )
  }

  const { commits, commitsThisYear, hasCommitsThisYear } = data

  return (
    <div className="mx-auto min-h-screen max-w-6xl pt-4 pb-32">
      <div className="px-6">
        <div className="mb-4 flex items-center">
          <Link to={'/'} className="flex items-center gap-2">
            <img
              src="/images/logo.png"
              alt="GitBack"
              className="size-10 object-cover"
            />
            <h3 className="text-2xl font-bold">GitBack</h3>
          </Link>
        </div>
        <div className="mb-10 flex flex-wrap items-center justify-between gap-3">
          <div className="">
            <h1 className="mb-1 text-3xl font-black tracking-tight lg:mb-3 lg:text-5xl">
              {repo}
            </h1>
            <h3 className="text-lg font-medium tracking-tight lg:text-xl">
              {username}
            </h3>
          </div>

          <a
            href={`https://github.com/${username}/${repo}`}
            target="_blank"
            rel="noopener noreferrer"
            className="bg-polar-sand text-obsidian-field flex items-center gap-2 rounded-full px-4 py-1.5 font-semibold"
          >
            <svg
              className="size-5"
              fill="currentColor"
              viewBox="0 0 24 24"
              aria-hidden="true"
            >
              <path
                fillRule="evenodd"
                d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"
                clipRule="evenodd"
              />
            </svg>
            <span className="">View on GitHub</span>
            <svg
              className="size-5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              aria-hidden="true"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
              />
            </svg>
          </a>
        </div>

        <div className="my-10">
          <CommitGraph
            commits={commits}
            totalContributors={data.totalContributors}
            totalAdded={data.totalAdded}
            totalRemoved={data.totalRemoved}
          />
        </div>

        <div className="my-10 grid grid-rows-3 space-y-3">
          <div className="flex gap-3 transition-all ease-in-out">
            <div className="bg-ion-drift flex-1 rounded-full transition-all duration-1000 ease-in-out hover:flex-2 hover:duration-100" />
            <div className="bg-core-flux flex-1 rounded-full transition-all duration-1000 ease-in-out hover:flex-2 hover:duration-100" />
            <div className="bg-pinky flex-1 rounded-full transition-all duration-1000 ease-in-out hover:flex-2 hover:duration-100" />
            <div className="bg-polar-sand flex-1 rounded-full transition-all duration-1000 ease-in-out hover:flex-2 hover:duration-100" />
            <div className="bg-pinky flex-1 rounded-full transition-all duration-1000 ease-in-out hover:flex-2 hover:duration-100" />
          </div>
          <div className="flex gap-3 transition-all ease-in-out">
            <div className="bg-pinky flex-1 rounded-full transition-all duration-1000 ease-in-out hover:flex-2 hover:duration-100" />
            <div className="bg-polar-sand flex-1 rounded-full transition-all duration-1000 ease-in-out hover:flex-2 hover:duration-100" />
            <div className="flex-center text-polar-sand bg-core-flux w-max flex-col rounded-full px-12 py-5 text-center transition-all duration-1000 ease-in-out hover:px-20 hover:duration-100">
              <p className="text-obsidian-field w-full text-5xl font-black">
                Git Wrapped
              </p>
            </div>
            <div className="bg-ion-drift flex-1 rounded-full transition-all duration-1000 ease-in-out hover:flex-2 hover:duration-100" />
            <div className="bg-polar-sand flex-1 rounded-full transition-all duration-1000 ease-in-out hover:flex-2 hover:duration-100" />
          </div>
          <div className="flex gap-3 transition-all ease-in-out">
            <div className="bg-core-flux flex-1 rounded-full transition-all duration-1000 ease-in-out hover:flex-2 hover:duration-100" />
            <div className="bg-ion-drift flex-1 rounded-full transition-all duration-1000 ease-in-out hover:flex-2 hover:duration-100" />
            <div className="bg-polar-sand flex-1 rounded-full transition-all duration-1000 ease-in-out hover:flex-2 hover:duration-100" />
            <div className="bg-core-flux flex-1 rounded-full transition-all duration-1000 ease-in-out hover:flex-2 hover:duration-100" />
            <div className="bg-pinky flex-1 rounded-full transition-all duration-1000 ease-in-out hover:flex-2 hover:duration-100" />
          </div>
        </div>

        {data.pullRequests && data.pullRequests.items.length > 0 && (
          <TopGitHubPRs prs={data.pullRequests.items} />
        )}

        {hasCommitsThisYear ? (
          <div className="mt-52 space-y-52">
            <CraziestWeek stats={commitsThisYear} />
            <TopContributors commits={commitsThisYear} />
            <FileCountDistribution commits={commitsThisYear} />
            <CommitWordCloud commits={commitsThisYear} />
            <BiggestCommits
              commits={commitsThisYear}
              repo={repo}
              username={username}
            />
            <CommitGrid commits={commitsThisYear} />
            {data.pullRequests && data.pullRequests.items.length > 0 && (
              <TopGitHubPRs prs={data.pullRequests.items} />
            )}

            <OverviewRecap
              commits={commitsThisYear}
              pullRequests={data.pullRequests}
              repoName={repo}
              username={username}
            />
          </div>
        ) : (
          <div className="mt-52 space-y-52">
            <p className="text-2xl font-semibold">
              No commits were made in the year of {THIS_YEAR}, boooooring...
            </p>
          </div>
        )}
      </div>
    </div>
  )
}
