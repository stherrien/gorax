import React, { useState } from 'react';
import {
  useRoles,
  useCreateRole,
  useUpdateRole,
  useDeleteRole,
  usePermissions,
  useUpdateRolePermissions,
} from '../hooks/useRoles';
import type { Role, Permission } from '../api/rbac';

export const RoleManagement: React.FC = () => {
  const { data: roles = [], isLoading } = useRoles();
  const { data: allPermissions = [] } = usePermissions();
  const [selectedRole, setSelectedRole] = useState<Role | null>(null);
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [showEditDialog, setShowEditDialog] = useState(false);
  const [showPermissionsDialog, setShowPermissionsDialog] = useState(false);

  if (isLoading) {
    return <div className="p-6">Loading roles...</div>;
  }

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">Role Management</h1>
        <button
          onClick={() => setShowCreateDialog(true)}
          className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
        >
          Create Role
        </button>
      </div>

      <RoleList
        roles={roles}
        onEdit={(role) => {
          setSelectedRole(role);
          setShowEditDialog(true);
        }}
        onDelete={(role) => {
          if (confirm(`Delete role "${role.name}"?`)) {
            // Will be handled by mutation
          }
        }}
        onManagePermissions={(role) => {
          setSelectedRole(role);
          setShowPermissionsDialog(true);
        }}
      />

      {showCreateDialog && (
        <RoleFormDialog
          onClose={() => setShowCreateDialog(false)}
          allPermissions={allPermissions}
        />
      )}

      {showEditDialog && selectedRole && (
        <RoleFormDialog
          role={selectedRole}
          onClose={() => {
            setShowEditDialog(false);
            setSelectedRole(null);
          }}
          allPermissions={allPermissions}
        />
      )}

      {showPermissionsDialog && selectedRole && (
        <PermissionMatrixDialog
          role={selectedRole}
          allPermissions={allPermissions}
          onClose={() => {
            setShowPermissionsDialog(false);
            setSelectedRole(null);
          }}
        />
      )}
    </div>
  );
};

interface RoleListProps {
  roles: Role[];
  onEdit: (role: Role) => void;
  onDelete: (role: Role) => void;
  onManagePermissions: (role: Role) => void;
}

const RoleList: React.FC<RoleListProps> = ({ roles, onEdit, onDelete: _onDelete, onManagePermissions }) => {
  const deleteRole = useDeleteRole();

  return (
    <div className="bg-white shadow rounded-lg overflow-hidden">
      <table className="min-w-full divide-y divide-gray-200">
        <thead className="bg-gray-50">
          <tr>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Name
            </th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Description
            </th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Permissions
            </th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Type
            </th>
            <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
              Actions
            </th>
          </tr>
        </thead>
        <tbody className="bg-white divide-y divide-gray-200">
          {roles.map((role) => (
            <tr key={role.id}>
              <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                {role.name}
              </td>
              <td className="px-6 py-4 text-sm text-gray-500">
                {role.description}
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                {role.permissions?.length || 0} permissions
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                {role.is_system ? (
                  <span className="px-2 py-1 text-xs bg-blue-100 text-blue-800 rounded">System</span>
                ) : (
                  <span className="px-2 py-1 text-xs bg-gray-100 text-gray-800 rounded">Custom</span>
                )}
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                <button
                  onClick={() => onManagePermissions(role)}
                  className="text-blue-600 hover:text-blue-900 mr-3"
                >
                  Permissions
                </button>
                {!role.is_system && (
                  <>
                    <button
                      onClick={() => onEdit(role)}
                      className="text-indigo-600 hover:text-indigo-900 mr-3"
                    >
                      Edit
                    </button>
                    <button
                      onClick={() => {
                        if (confirm(`Delete role "${role.name}"?`)) {
                          deleteRole.mutate(role.id);
                        }
                      }}
                      className="text-red-600 hover:text-red-900"
                    >
                      Delete
                    </button>
                  </>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

interface RoleFormDialogProps {
  role?: Role;
  onClose: () => void;
  allPermissions: Permission[];
}

const RoleFormDialog: React.FC<RoleFormDialogProps> = ({ role, onClose, allPermissions }) => {
  const [name, setName] = useState(role?.name || '');
  const [description, setDescription] = useState(role?.description || '');
  const [selectedPermissions, setSelectedPermissions] = useState<string[]>(
    role?.permissions?.map((p) => p.id) || []
  );

  const createRole = useCreateRole();
  const updateRole = useUpdateRole(role?.id || '');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (role) {
      await updateRole.mutateAsync({
        name,
        description,
        permission_ids: selectedPermissions,
      });
    } else {
      await createRole.mutateAsync({
        name,
        description,
        permission_ids: selectedPermissions,
      });
    }

    onClose();
  };

  const togglePermission = (permId: string) => {
    setSelectedPermissions((prev) =>
      prev.includes(permId) ? prev.filter((id) => id !== permId) : [...prev, permId]
    );
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 w-full max-w-2xl max-h-[90vh] overflow-y-auto">
        <h2 className="text-xl font-bold mb-4">
          {role ? 'Edit Role' : 'Create Role'}
        </h2>

        <form onSubmit={handleSubmit}>
          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Name
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
              required
            />
          </div>

          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Description
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
              rows={3}
            />
          </div>

          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Permissions
            </label>
            <div className="border border-gray-300 rounded p-3 max-h-60 overflow-y-auto">
              {groupPermissionsByResource(allPermissions).map(([resource, perms]) => (
                <div key={resource} className="mb-3">
                  <div className="font-medium text-sm text-gray-700 mb-1">
                    {resource}
                  </div>
                  <div className="space-y-1 ml-4">
                    {perms.map((perm) => (
                      <label key={perm.id} className="flex items-center">
                        <input
                          type="checkbox"
                          checked={selectedPermissions.includes(perm.id)}
                          onChange={() => togglePermission(perm.id)}
                          className="mr-2"
                        />
                        <span className="text-sm text-gray-600">
                          {perm.action} - {perm.description}
                        </span>
                      </label>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className="flex justify-end space-x-3">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-gray-700 border border-gray-300 rounded hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
            >
              {role ? 'Update' : 'Create'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

interface PermissionMatrixDialogProps {
  role: Role;
  allPermissions: Permission[];
  onClose: () => void;
}

const PermissionMatrixDialog: React.FC<PermissionMatrixDialogProps> = ({
  role,
  allPermissions,
  onClose,
}) => {
  const [selectedPermissions, setSelectedPermissions] = useState<string[]>(
    role.permissions?.map((p) => p.id) || []
  );
  const updatePermissions = useUpdateRolePermissions(role.id);

  const handleSave = async () => {
    await updatePermissions.mutateAsync(selectedPermissions);
    onClose();
  };

  const togglePermission = (permId: string) => {
    setSelectedPermissions((prev) =>
      prev.includes(permId) ? prev.filter((id) => id !== permId) : [...prev, permId]
    );
  };

  const groupedPermissions = groupPermissionsByResource(allPermissions);

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 w-full max-w-4xl max-h-[90vh] overflow-y-auto">
        <h2 className="text-xl font-bold mb-4">
          Manage Permissions for "{role.name}"
        </h2>

        <div className="border border-gray-300 rounded overflow-hidden mb-4">
          <table className="min-w-full">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">
                  Resource
                </th>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {groupedPermissions.map(([resource, perms]) => (
                <tr key={resource}>
                  <td className="px-4 py-3 font-medium text-sm text-gray-900">
                    {resource}
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex flex-wrap gap-2">
                      {perms.map((perm) => (
                        <label
                          key={perm.id}
                          className="flex items-center px-3 py-1 border border-gray-300 rounded cursor-pointer hover:bg-gray-50"
                        >
                          <input
                            type="checkbox"
                            checked={selectedPermissions.includes(perm.id)}
                            onChange={() => togglePermission(perm.id)}
                            className="mr-2"
                            disabled={role.is_system}
                          />
                          <span className="text-sm">{perm.action}</span>
                        </label>
                      ))}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {role.is_system && (
          <div className="mb-4 p-3 bg-yellow-50 border border-yellow-200 rounded text-sm text-yellow-800">
            System roles cannot be modified. You can view permissions but cannot change them.
          </div>
        )}

        <div className="flex justify-end space-x-3">
          <button
            onClick={onClose}
            className="px-4 py-2 text-gray-700 border border-gray-300 rounded hover:bg-gray-50"
          >
            {role.is_system ? 'Close' : 'Cancel'}
          </button>
          {!role.is_system && (
            <button
              onClick={handleSave}
              className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
            >
              Save Changes
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

// Helper function to group permissions by resource
function groupPermissionsByResource(permissions: Permission[]): [string, Permission[]][] {
  const grouped = permissions.reduce((acc, perm) => {
    if (!acc[perm.resource]) {
      acc[perm.resource] = [];
    }
    acc[perm.resource].push(perm);
    return acc;
  }, {} as Record<string, Permission[]>);

  return Object.entries(grouped).sort(([a], [b]) => a.localeCompare(b));
}

export default RoleManagement;
