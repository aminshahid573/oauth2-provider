document.addEventListener('DOMContentLoaded', () => {
    // Client page logic
    const clientsTableBody = document.getElementById('clientsTableBody');
    if (clientsTableBody) {
        loadClients();
        setupClientPageEventListeners();
    }

    // User page logic
    const usersTableBody = document.getElementById('usersTableBody');
    if (usersTableBody) {
        loadUsers();
        setupUserPageEventListeners();
    }
});

// --- USER MANAGEMENT FUNCTIONS ---

function setupUserPageEventListeners() {
    const addUserBtn = document.getElementById('addUserBtn');
    const userModal = document.getElementById('userModal');
    const closeModalBtn = document.getElementById('closeModalBtn');
    const cancelModalBtn = document.getElementById('cancelModalBtn');
    const userForm = document.getElementById('userForm');

    addUserBtn.addEventListener('click', () => userModal.classList.remove('hidden'));
    closeModalBtn.addEventListener('click', () => userModal.classList.add('hidden'));
    cancelModalBtn.addEventListener('click', () => userModal.classList.add('hidden'));
    userForm.addEventListener('submit', handleUserFormSubmit);
}

async function handleUserFormSubmit(event) {
    event.preventDefault();
    const form = event.target;
    const formData = new FormData(form);
    const csrfToken = form.querySelector('input[name="_csrf"]').value;

    const payload = {
        username: formData.get('username'),
        password: formData.get('password'),
        role: formData.get('role'),
    };

    try {
        const response = await fetch('/api/admin/users', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': csrfToken
            },
            body: JSON.stringify(payload)
        });
        const result = await response.json();
        if (!response.ok) throw new Error(result.error_description || 'Failed to create user.');

        document.getElementById('userModal').classList.add('hidden');
        showNotification(`User "${result.username}" created successfully.`, 'success');
        loadUsers();
    } catch (error) {
        console.error('Error creating user:', error);
        showNotification(error.message, 'error');
    }
}

async function loadUsers() {
    const tableBody = document.getElementById('usersTableBody');
    try {
        const response = await fetch('/api/admin/users');
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
        const users = await response.json();
        tableBody.innerHTML = '';
        if (!users || users.length === 0) {
            tableBody.innerHTML = '<tr><td colspan="4" class="text-center">No users found.</td></tr>';
            return;
        }
        users.forEach(user => {
            const row = document.createElement('tr');
            row.innerHTML = `
                <td>${escapeHTML(user.username)}</td>
                <td>${escapeHTML(user.role)}</td>
                <td><code>${escapeHTML(user.id)}</code></td>
                <td class="action-buttons">
                    <a href="#" class="edit-btn" data-user-id="${escapeHTML(user.id)}">Edit</a>
                    <a href="#" class="delete-btn" data-user-id="${escapeHTML(user.id)}">Delete</a>
                </td>
            `;
            tableBody.appendChild(row);
        });
    } catch (error) {
        console.error('Failed to load users:', error);
        tableBody.innerHTML = '<tr><td colspan="4" class="text-center">Failed to load users.</td></tr>';
    }
}

// --- CLIENT MANAGEMENT FUNCTIONS ---

function setupClientPageEventListeners() {
    const addClientBtn = document.getElementById('addClientBtn');
    const clientModal = document.getElementById('clientModal');
    const closeModalBtn = document.getElementById('closeModalBtn');
    const cancelModalBtn = document.getElementById('cancelModalBtn');
    const clientForm = document.getElementById('clientForm');
    const secretModal = document.getElementById('secretModal');
    const closeSecretModalBtn = document.getElementById('closeSecretModalBtn');

    // Modal open/close
    addClientBtn.addEventListener('click', () => {
        clientForm.reset();
        clientForm.removeAttribute('data-editing-client-id');
        document.getElementById('modalTitle').textContent = 'Add New Client';
        clientModal.classList.remove('hidden');
    });
    closeModalBtn.addEventListener('click', () => clientModal.classList.add('hidden'));
    cancelModalBtn.addEventListener('click', () => clientModal.classList.add('hidden'));
    closeSecretModalBtn.addEventListener('click', () => secretModal.classList.add('hidden'));

    // Form submission (handles both create and update)
    clientForm.addEventListener('submit', handleClientFormSubmit);

    // Event delegation for edit and delete buttons
    document.getElementById('clientsTableBody').addEventListener('click', (event) => {
        if (event.target.classList.contains('delete-btn')) {
            event.preventDefault();
            const clientID = event.target.dataset.clientId;
            if (confirm(`Are you sure you want to delete client ${clientID}?`)) {
                deleteClient(clientID);
            }
        }
        if (event.target.classList.contains('edit-btn')) {
            event.preventDefault();
            const clientID = event.target.dataset.clientId;
            handleEditClick(clientID);
        }
    });
}

async function handleEditClick(clientID) {
    try {
        const response = await fetch(`/api/admin/clients/${clientID}`);
        if (!response.ok) throw new Error('Failed to fetch client details.');
        const client = await response.json();

        const form = document.getElementById('clientForm');
        form.reset();
        form.setAttribute('data-editing-client-id', client.client_id);

        document.getElementById('modalTitle').textContent = `Edit Client: ${client.name}`;
        form.elements.name.value = client.name;
        form.elements.redirect_uris.value = client.redirect_uris.join('\n');
        form.elements.scopes.value = client.scopes.join(' ');

        // Check the correct checkboxes for grant_types and response_types
        client.grant_types.forEach(type => {
            const checkbox = form.querySelector(`input[name="grant_types"][value="${type}"]`);
            if (checkbox) checkbox.checked = true;
        });
        client.response_types.forEach(type => {
            const checkbox = form.querySelector(`input[name="response_types"][value="${type}"]`);
            if (checkbox) checkbox.checked = true;
        });

        document.getElementById('clientModal').classList.remove('hidden');
    } catch (error) {
        console.error('Error preparing edit form:', error);
        showNotification(error.message, 'error');
    }
}

async function handleClientFormSubmit(event) {
    event.preventDefault();
    const form = event.target;
    const editingClientID = form.dataset.editingClientId;
    const method = editingClientID ? 'PUT' : 'POST';
    const url = editingClientID ? `/api/admin/clients/${editingClientID}` : '/api/admin/clients';

    const formData = new FormData(form);
    const csrfToken = form.querySelector('input[name="_csrf"]').value;

    const payload = {
        name: formData.get('name'),
        redirect_uris: formData.get('redirect_uris').split('\n').map(uri => uri.trim()).filter(uri => uri),
        grant_types: formData.getAll('grant_types'),
        response_types: formData.getAll('response_types'),
        scopes: formData.get('scopes').split(' ').map(s => s.trim()).filter(s => s),
    };

    try {
        const response = await fetch(url, {
            method: method,
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': csrfToken
            },
            body: JSON.stringify(payload)
        });

        const result = await response.json();
        if (!response.ok) throw new Error(result.error_description || `Failed to ${editingClientID ? 'update' : 'create'} client.`);

        document.getElementById('clientModal').classList.add('hidden');
        
        if (method === 'POST') {
            showNotification(`Client "${result.name}" created successfully.`, 'success');
            showSecretModal(result.client_id, result.client_secret);
        } else {
            showNotification(`Client "${result.name}" updated successfully.`, 'success');
        }
        loadClients();

    } catch (error) {
        console.error(`Error ${editingClientID ? 'updating' : 'creating'} client:`, error);
        showNotification(error.message, 'error');
    }
}

// ... (deleteClient, showSecretModal, showNotification, loadClients, escapeHTML functions remain the same)
async function deleteClient(clientID) {
    const csrfToken = document.querySelector('input[name="_csrf"]').value;
    try {
        const response = await fetch(`/api/admin/clients/${clientID}`, {
            method: 'DELETE',
            headers: { 'X-CSRF-Token': csrfToken }
        });
        if (!response.ok) {
            const result = await response.json().catch(() => ({}));
            throw new Error(result.error_description || 'Failed to delete client.');
        }
        showNotification(`Client ${clientID} deleted successfully.`, 'success');
        loadClients();
    } catch (error) {
        console.error('Error deleting client:', error);
        showNotification(error.message, 'error');
    }
}

function showSecretModal(clientID, clientSecret) {
    document.getElementById('newClientId').textContent = clientID;
    document.getElementById('newClientSecret').textContent = clientSecret;
    document.getElementById('secretModal').classList.remove('hidden');
}

function showNotification(message, type) {
    const notification = document.getElementById('notification');
    notification.className = `notification ${type}`;
    notification.textContent = message;
    notification.classList.remove('hidden');
    setTimeout(() => { notification.classList.add('hidden'); }, 5000);
}

async function loadClients() {
    const tableBody = document.getElementById('clientsTableBody');
    try {
        const response = await fetch('/api/admin/clients');
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
        const clients = await response.json();
        tableBody.innerHTML = '';
        if (clients.length === 0) {
            tableBody.innerHTML = '<tr><td colspan="4" class="text-center">No clients found.</td></tr>';
            return;
        }
        clients.forEach(client => {
            const row = document.createElement('tr');
            const grantTypes = client.grant_types ? client.grant_types.join(', ') : 'N/A';
            row.innerHTML = `
                <td>${escapeHTML(client.name)}</td>
                <td><code>${escapeHTML(client.client_id)}</code></td>
                <td>${escapeHTML(grantTypes)}</td>
                <td class="action-buttons">
                    <a href="#" class="edit-btn" data-client-id="${escapeHTML(client.client_id)}">Edit</a>
                    <a href="#" class="delete-btn" data-client-id="${escapeHTML(client.client_id)}">Delete</a>
                </td>
            `;
            tableBody.appendChild(row);
        });
    } catch (error) {
        console.error('Failed to load clients:', error);
        tableBody.innerHTML = '<tr><td colspan="4" class="text-center">Failed to load clients. See console for details.</td></tr>';
    }
}

function escapeHTML(str) {
    if (str === null || str === undefined) return '';
    return str.toString().replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;').replace(/'/g, '&#039;');
}