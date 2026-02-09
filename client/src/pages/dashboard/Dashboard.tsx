import { useAuthStore } from "@/store/auth.store"

const Dashboard = () => {

  const { user } = useAuthStore((state) => state);

  console.log(user);

  return (
    <div>Dashboard</div>
  )
}

export default Dashboard