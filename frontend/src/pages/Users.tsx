import { useEffect, useState } from 'react';
import { apiClient } from '../api/client';
import type { User, UserRole } from '../types';
import Layout from '../components/Layout';
import { extractErrorMessage } from '../utils/errors';

export default function Users() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [updatingId, setUpdatingId] = useState<number | null>(null);

  useEffect(() => {
    fetchUsers();
  }, []);

  const fetchUsers = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await apiClient.getUsers();
      setUsers(data ?? []);
    } catch (err) {
      setError(extractErrorMessage(err, 'Failed to load users'));
    } finally {
      setLoading(false);
    }
  };

  const handleRoleChange = async (user: User, newRole: UserRole) => {
    if (user.role === 'ADMIN' || user.role === newRole) return;
    try {
      setUpdatingId(user.id);
      await apiClient.updateUserRole(user.id, newRole);
      setUsers((prev) =>
        prev.map((u) => (u.id === user.id ? { ...u, role: newRole } : u))
      );
    } catch (err) {
      alert(extractErrorMessage(err, 'Failed to update role'));
    } finally {
      setUpdatingId(null);
    }
  };

  const getRoleBadgeColor = (role: UserRole) => {
    switch (role) {
      case 'ADMIN':
        return 'bg-gradient-to-r from-purple-600 to-purple-700 text-white';
      case 'VIP':
        return 'bg-gradient-to-r from-amber-500 to-amber-600 text-white';
      default:
        return 'bg-[#1e222d] text-[#d1d4dc] border border-[#2a2e39]';
    }
  };

  const getRoleIcon = (role: UserRole) => {
    switch (role) {
      case 'ADMIN':
        return '👑';
      case 'VIP':
        return '⭐';
      default:
        return '👤';
    }
  };

  return (
    <Layout>
      {/* Main Content */}
      <div className="container mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {loading && (
          <div className="flex items-center justify-center py-20">
            <div className="flex flex-col items-center gap-3">
              <div className="w-8 h-8 border-4 border-[#26a69a] border-t-transparent rounded-full animate-spin"></div>
              <div className="text-[#d1d4dc] text-sm font-medium">Loading users...</div>
            </div>
          </div>
        )}

        {error && (
          <div className="mb-6 bg-[#ef5350] bg-opacity-90 backdrop-blur-sm text-white p-4 rounded-lg border border-[#ef5350] flex items-center justify-between">
            <span>{error}</span>
            <button
              onClick={fetchUsers}
              className="ml-4 px-3 py-1 text-sm bg-white bg-opacity-20 hover:bg-opacity-30 rounded transition-all"
            >
              Retry
            </button>
          </div>
        )}

        {!loading && !error && (
          <div className="bg-[#131722] border border-[#2a2e39] rounded-lg shadow-2xl overflow-hidden">
            {/* Table Header */}
            <div className="px-6 py-4 border-b border-[#2a2e39] bg-gradient-to-r from-[#131722] to-[#1a1e2e]">
              <div className="flex items-center justify-between">
                <h2 className="text-lg font-semibold text-[#d1d4dc]">
                  Users ({users.length})
                </h2>
                <div className="flex items-center gap-4 text-xs text-[#758696]">
                  <div className="flex items-center gap-1">
                    <span className="w-2 h-2 rounded-full bg-purple-500"></span>
                    <span>Admin</span>
                  </div>
                  <div className="flex items-center gap-1">
                    <span className="w-2 h-2 rounded-full bg-amber-500"></span>
                    <span>VIP</span>
                  </div>
                  <div className="flex items-center gap-1">
                    <span className="w-2 h-2 rounded-full bg-[#2a2e39]"></span>
                    <span>Normal</span>
                  </div>
                </div>
              </div>
            </div>

            {/* Table */}
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-[#2a2e39]">
                <thead className="bg-[#1e222d]">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-semibold text-[#d1d4dc] uppercase tracking-wider">
                      User
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-semibold text-[#d1d4dc] uppercase tracking-wider">
                      Email
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-semibold text-[#d1d4dc] uppercase tracking-wider">
                      Role
                    </th>
                    <th className="px-6 py-3 text-right text-xs font-semibold text-[#d1d4dc] uppercase tracking-wider">
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-[#131722] divide-y divide-[#2a2e39]">
                  {users.map((user) => (
                    <tr key={user.id} className="hover:bg-[#1a1e2e] transition-colors">
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="flex items-center gap-3">
                          <div className="w-10 h-10 rounded-full bg-gradient-to-br from-[#26a69a] to-[#1e7a6e] flex items-center justify-center text-white font-bold">
                            {user.email.charAt(0).toUpperCase()}
                          </div>
                          <div>
                            <div className="text-sm font-medium text-[#d1d4dc]">
                              {[user.first_name, user.last_name].filter(Boolean).join(' ') || 'No name'}
                            </div>
                            <div className="text-xs text-[#758696]">
                              ID: {user.id}
                            </div>
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm text-[#d1d4dc]">{user.email}</div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span
                          className={`inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold rounded-full ${getRoleBadgeColor(user.role)}`}
                        >
                          <span>{getRoleIcon(user.role)}</span>
                          <span>{user.role}</span>
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-right">
                        {user.role === 'ADMIN' ? (
                          <span className="text-xs text-[#758696] italic">Protected</span>
                        ) : (
                          <select
                            value={user.role}
                            onChange={(e) =>
                              handleRoleChange(user, e.target.value as UserRole)
                            }
                            disabled={updatingId === user.id}
                            className="px-3 py-1.5 text-sm bg-[#1e222d] border border-[#2a2e39] rounded-lg text-[#d1d4dc] focus:outline-none focus:ring-2 focus:ring-[#26a69a] hover:bg-[#252936] disabled:opacity-50 disabled:cursor-not-allowed transition-all cursor-pointer"
                          >
                            <option value="NORMAL" className="bg-[#1e222d] text-[#d1d4dc]">NORMAL</option>
                            <option value="VIP" className="bg-[#1e222d] text-[#d1d4dc]">VIP</option>
                          </select>
                        )}
                        {updatingId === user.id && (
                          <div className="inline-flex items-center gap-1 ml-2 text-xs text-[#758696]">
                            <div className="w-3 h-3 border-2 border-[#26a69a] border-t-transparent rounded-full animate-spin"></div>
                            <span>Updating...</span>
                          </div>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {users.length === 0 && (
              <div className="text-center py-12">
                <div className="inline-block p-4 bg-[#1e222d] rounded-full mb-4">
                  <span className="text-4xl">👥</span>
                </div>
                <p className="text-[#758696]">No users found</p>
              </div>
            )}
          </div>
        )}
      </div>
    </Layout>
  );
}
