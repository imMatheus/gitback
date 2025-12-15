import React from 'react'

export const LoadingAnimation: React.FC = ({}) => {
  const animationDuration = 2100
  return (
    <>
      <style>{`
        @keyframes loader-1 {
            0%, 50%, 100% { width: 50%; height: 50%; }
            25% { width: 33%; height: 50%; }
            75% { width: 50%; height: 67%; }
        }
        @keyframes loader-2 {
            0%, 50%, 100% { width: 50%; height: 50%; }
            25% { width: 67%; height: 50%; }
            75% { width: 50%; height: 33%; }
        }
        `}</style>
      <div className="">
        <div className="h-52 w-52 relative">
          <div
            className="absolute size-1/2 bg-core-flux rounded-full top-0 left-0"
            style={{ animation: `loader-1 ${animationDuration}ms infinite` }}
          />
          <div
            className="absolute size-1/2 bg-pinky rounded-full top-0 right-0"
            style={{ animation: `loader-2 ${animationDuration}ms infinite` }}
          />
          <div
            className="absolute size-1/2 bg-ion-drift rounded-full bottom-0 left-0"
            style={{ animation: `loader-2 ${animationDuration}ms infinite` }}
          />
          <div
            className="absolute size-1/2 bg-magnetic-mist rounded-full bottom-0 right-0"
            style={{ animation: `loader-1 ${animationDuration}ms infinite` }}
          />
        </div>
        <p className="text-sm text-center font-medium pt-2">
          cranking the numbers...
        </p>
      </div>
    </>
  )
}
