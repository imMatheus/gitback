import { cn } from '@/lib/utils'
import type { CommitStats } from '@/types'
import { useMemo } from 'react'

interface CommitWordCloudProps {
  commits: CommitStats[]
}

export const CommitWordCloud: React.FC<CommitWordCloudProps> = ({
  commits,
}) => {
  const { words, wordCounts } = useMemo(() => {
    const wordCounts = {
      wtf: 0,
      fixme: 0,
      todo: 0,
      hack: 0,
      test: 0,
      please: 0,
    }

    for (const commit of commits) {
      const message = commit.message.toLowerCase().split(' ')
      if (message.includes('wtf')) {
        wordCounts.wtf++
      } else if (message.includes('fixme')) {
        wordCounts.fixme++
      } else if (message.includes('todo')) {
        wordCounts.todo++
      } else if (message.includes('hack')) {
        wordCounts.hack++
      } else if (message.includes('test')) {
        wordCounts.test++
      } else if (message.includes('please')) {
        wordCounts.please++
      }
    }

    const messages = commits.map((commit) => commit.message)
    return { words: extractWords(messages).slice(0, 20), wordCounts }
  }, [commits])

  return (
    <div className="">
      <h3 className="mb-4 text-6xl font-black">Bänger commits!</h3>
      <p className="mb-8 text-xl font-semibold">
        coming up with commit messages is hard, you did not do great :){' '}
      </p>

      <div className="text-obsidian-field grid grid-cols-2 gap-2">
        <div className="bg-alloy-ember flex-1 rounded-full px-10 py-6">
          <p className="text-5xl font-bold">{wordCounts.wtf} "WTF"</p>
          <p className="text-xl font-bold">
            {wordCounts.wtf > 0
              ? 'me when i give up'
              : 'swearing is rude, so makes sense tbh'}
          </p>
        </div>

        <div className="bg-pinky flex-1 rounded-full px-10 py-6">
          <p className="text-5xl font-bold">{wordCounts.fixme} "FIXME"</p>
          <p className="text-xl font-bold">
            {wordCounts.fixme > 0
              ? 'we all know none of these where fixed'
              : 'someone is NOT a perfectionist'}
          </p>
        </div>

        <div className="bg-ion-drift flex-1 rounded-full px-10 py-6">
          <p className="text-5xl font-bold">{wordCounts.todo} "TODO"</p>
          <p className="text-xl font-bold">
            {wordCounts.todo > 0 ? (
              <>
                how about you <i className="font-black">do</i> some pushups{' '}
              </>
            ) : (
              'someone is NOT a procrastinator'
            )}
          </p>
        </div>

        <div className="bg-core-flux flex-1 rounded-full px-10 py-6">
          <p className="text-5xl font-bold">{wordCounts.hack} "HACK"</p>
          <p className="text-xl font-bold">
            {wordCounts.hack > 0
              ? '"it works, i sweeear" ahh'
              : 'ok MR.perfectcodesoyboy'}
          </p>
        </div>

        <div className="bg-polar-sand flex-1 rounded-full px-10 py-6">
          <p className="text-5xl font-bold">{wordCounts.test} "TEST"</p>
          <p className="text-xl font-bold">
            this SCREAMS "i dont trust my code"
          </p>
        </div>

        <div className="bg-alloy-ember flex-1 rounded-full px-10 py-6">
          <p className="text-5xl font-bold">{wordCounts.please} "PLEASE"</p>
          <p className="text-xl font-bold">
            {wordCounts.please > 0
              ? '"whats the magic word" ahh commits'
              : 'someone is not polite lmao'}
          </p>
        </div>
      </div>

      <div className="mt-5 flex flex-wrap gap-2">
        {words.length > 0 &&
          words.map((word, index) => (
            <div
              key={word.text}
              className={cn(
                'text-obsidian-field max-w-max flex-1 rounded-full px-5 py-2',
                {
                  'bg-alloy-ember': index % 5 === 0,
                  'bg-pinky': index % 5 === 1,
                  'bg-ion-drift': index % 5 === 2,
                  'bg-core-flux': index % 5 === 3,
                  'bg-polar-sand': index % 5 === 4,
                }
              )}
            >
              <p className="truncate text-2xl font-bold whitespace-nowrap">
                {word.appearances} "{word.text}"
              </p>
            </div>
          ))}
      </div>
    </div>
  )
}

function extractWords(messages: string[]): {
  text: string
  appearances: number
}[] {
  const freqMap: Record<string, number> = {}

  for (const message of messages) {
    // Split by whitespace and punctuation, convert to lowercase
    const words = message
      .toLowerCase()
      .replace(/[^\w\s]/g, ' ')
      .split(/\s+/)
      .filter((word) => word.length > 2)

    for (const word of words) {
      freqMap[word] = (freqMap[word] || 0) + 1
    }
  }

  // Convert to array and filter out words that appear only once
  return Object.entries(freqMap)
    .filter(([_, count]) => count > 1)
    .map(([word, count]) => ({ text: word, appearances: count }))
    .sort((a, b) => b.appearances - a.appearances)
  // .slice(0, 2090) // Limit to top 100 words
}
