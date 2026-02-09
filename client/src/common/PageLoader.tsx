import { Loader } from "./Loader"

const PageLoader = ({className="w-full h-screen flex items-center justify-center"} : {className?: string}) => {
  return (
    <div className={className}>
        <Loader className="w-8 h-8 text-indigo-500" />
    </div>
  )
}

export default PageLoader