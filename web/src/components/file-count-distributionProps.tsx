import React from 'react'
import type { CommitStats } from '@/types'
import { cn } from '@/lib/utils'

interface FileCountDistributionProps {
  commits: CommitStats[]
}

interface DistributionBarProps {
  label: string
  value: number
  maxValue: number
  index: number
}

const maxHeight = '400px'

const ALL_BARS = ['1', '2-15', '16-30', '31-70', '71-200', '200+']

const DistributionBar: React.FC<DistributionBarProps> = ({
  label,
  value,
  maxValue,
  index,
}) => {
  const indexD = index % 5
  return (
    <div className="flex h-full flex-col">
      <div className="flex flex-1 flex-col justify-end">
        <div
          className={cn(
            `text-obsidian-field flex-center rounded-full py-0.5 text-lg font-bold`,
            {
              'bg-ion-drift': indexD === 0,
              'bg-alloy-ember': indexD === 1,
              'bg-core-flux': indexD === 2,
              'bg-pinky': indexD === 3,
              'bg-polar-sand': indexD === 4,
            },
            {
              'py-4 text-5xl': index === 0,
              'py-3 text-4xl': index === 1,
              'py-2 text-3xl': index === 2,
              'py-1 text-2xl': index === 3,
            }
          )}
          style={{
            height:
              value === maxValue ? maxHeight : `${(value / maxValue) * 100}%`,
            minHeight: 'fit-content',
          }}
        >
          {value}
        </div>
      </div>
      <p className="mt-3 shrink-0 text-center text-xl font-bold">{label}</p>
    </div>
  )
}

export const FileCountDistribution: React.FC<FileCountDistributionProps> = ({
  commits,
}) => {
  const distribution = {
    '1': 0,
    '2-15': 0,
    '16-30': 0,
    '31-70': 0,
    '71-200': 0,
    '200+': 0,
  }

  for (const commit of commits) {
    if (commit.filesTouchedCount === 1) {
      distribution['1']++
    } else if (
      commit.filesTouchedCount >= 2 &&
      commit.filesTouchedCount <= 15
    ) {
      distribution['2-15']++
    } else if (
      commit.filesTouchedCount >= 16 &&
      commit.filesTouchedCount <= 30
    ) {
      distribution['16-30']++
    } else if (
      commit.filesTouchedCount >= 31 &&
      commit.filesTouchedCount <= 70
    ) {
      distribution['31-70']++
    } else if (
      commit.filesTouchedCount >= 71 &&
      commit.filesTouchedCount <= 200
    ) {
      distribution['71-200']++
    } else {
      distribution['200+']++
    }
  }

  const maxValue = Math.max(...Object.values(distribution))

  const bars = ALL_BARS.sort((a, b) => {
    const aValue = distribution[a as keyof typeof distribution]
    const bValue = distribution[b as keyof typeof distribution]
    return bValue - aValue
  })

  return (
    <div>
      <h3 className="my-4 text-6xl font-black">Commit file distribution</h3>
      <p className="mt-2 text-xl leading-relaxed font-semibold">
        This is the distribution of the number of files touched in a commit.
      </p>

      <div className="mt-5 grid grid-cols-6 gap-x-2">
        {bars.map((bar, index) => (
          <DistributionBar
            key={bar}
            label={bar}
            value={distribution[bar as keyof typeof distribution]}
            maxValue={maxValue}
            index={index}
          />
        ))}
      </div>
      <p className="mt-5 text-center text-xs font-semibold">
        the number of commits with certain number of files touched
      </p>
    </div>
  )
}
