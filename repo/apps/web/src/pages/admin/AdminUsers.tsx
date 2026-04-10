import { useQuery } from "@tanstack/react-query";
import api from "@/lib/api";

interface AdminUser {
  id: string;
  email: string;
  status: string;
  display_name: string;
}

interface UserListResponse {
  items: AdminUser[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export default function AdminUsers() {
  const { data, isLoading } = useQuery<UserListResponse>({
    queryKey: ["admin-users"],
    queryFn: async () => {
      const { data } = await api.get("/admin/users");
      return data;
    },
  });

  const users = data?.items ?? [];

  return (
    <div className="rounded-xl border border-gray-200 bg-white">
      <div className="border-b border-gray-200 px-6 py-4">
        <h2 className="text-lg font-semibold text-gray-900">User Management</h2>
      </div>
      {isLoading ? (
        <div className="p-6 text-center text-gray-500">Loading users...</div>
      ) : (
        <table className="w-full text-left text-sm">
          <thead className="border-b border-gray-200 bg-gray-50">
            <tr>
              <th className="px-6 py-3 font-medium text-gray-600">Name</th>
              <th className="px-6 py-3 font-medium text-gray-600">Email</th>
              <th className="px-6 py-3 font-medium text-gray-600">Status</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {users.map((u) => (
              <tr key={u.id} className="hover:bg-gray-50">
                <td className="px-6 py-3 font-medium text-gray-900">{u.display_name || "\u2014"}</td>
                <td className="px-6 py-3 text-gray-700">{u.email}</td>
                <td className="px-6 py-3">
                  <span className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${
                    u.status === "active" ? "bg-green-100 text-green-700" : "bg-red-100 text-red-700"
                  }`}>{u.status}</span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
