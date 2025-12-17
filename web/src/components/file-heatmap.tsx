import { cn } from '@/lib/utils'
import type { FileTouchCount } from '@/types'
import React from 'react'

interface FileHeatmapProps {
  mostTouchedFiles: FileTouchCount[]
}

export const FileHeatmap: React.FC<FileHeatmapProps> = ({
  mostTouchedFiles,
}) => {
  return (
    <div className="">
      <h3 className="my-4 text-6xl font-black">Most Touched Files</h3>
      <p className="mt-2 text-xl leading-relaxed font-semibold">
        These people lowkey carried, collectively pushing{' '}
      </p>

      <div className="mt-10 grid grid-flow-row-dense grid-cols-8 gap-2">
        {mostTouchedFiles.slice(0, 40).map((file, index) => (
          <div
            key={file.file}
            className={cn(
              'bg-pinky text-obsidian-field flex flex-col justify-center rounded-full px-4 py-2 pr-2! transition-all',
              index === 0
                ? 'col-span-3 row-span-2 p-10'
                : index === 1
                  ? 'col-span-3 col-start-2 row-start-3'
                  : index === 2 || index < 5 || index === 7
                    ? 'col-span-2'
                    : 'col-span-1',
              {
                'bg-core-flux': index % 5 === 0,
                'bg-ion-drift': index % 5 === 1,
                'bg-pinky': index % 5 === 2,
                'bg-polar-sand': index % 5 === 3,
                'bg-alloy-ember': index % 5 === 4,
              }
            )}
          >
            <p
              className={cn('truncate text-xs font-bold', {
                'text-4xl': index === 0,
                'text-3xl': index === 1,
                'text-2xl': index === 2,
                'text-xl': index === 3,
              })}
              title={file.file}
              style={{
                lineHeight: '1.4',
              }}
            >
              {file.file}
            </p>
            <p
              className={cn('text-xs font-semibold', {
                'text-xl font-bold': index === 0,
                'text-lg': index === 1,
              })}
            >
              {file.count} commits
            </p>
          </div>
        ))}
      </div>
    </div>
  )
}
