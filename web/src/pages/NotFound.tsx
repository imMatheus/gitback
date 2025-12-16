import { Link } from 'react-router'

export default function NotFound({ isRepo }: { isRepo: boolean }) {
  return (
    <div className="flex h-screen w-full items-center justify-center">
      <div className="text-center">
        <h1 className="mb-12 space-x-0 text-8xl font-bold">
          <span className="bg-pinky text-obsidian-field rounded-full px-3 py-1.5">
            4
          </span>
          <span className="bg-alloy-ember text-obsidian-field rounded-full px-3 py-1.5">
            0
          </span>
          <span className="bg-ion-drift text-obsidian-field rounded-full px-3 py-1.5">
            4
          </span>
        </h1>
        <p className="text-4xl font-bold">
          {isRepo ? 'could not find repository!' : 'could not find page!'}
        </p>

        <Link
          to="/"
          className="bg-polar-sand text-obsidian-field group mx-auto mt-12 flex w-max items-center justify-between gap-2 rounded-full p-6 text-2xl font-bold"
        >
          <p className="shrink-0 transition-all duration-1000 ease-in-out group-hover:-translate-x-1 group-hover:duration-100">
            ←
          </p>
          <p className="truncate">Go home!</p>
        </Link>
      </div>
    </div>
  )
}
