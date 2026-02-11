const API_BASE = 'http://localhost:8080';

const headers = {
  'Accept': 'application/vnd.api+json',
  'Content-Type': 'application/vnd.api+json',
};

// Map JSON:API source.pointer to form field id (without "data/attributes/" prefix).
// e.g. /data/attributes/name -> store-name or pet-name
function pointerToFieldId(pointer, resource) {
  if (!pointer) return null;
  if (pointer.startsWith('/data/attributes/')) {
    const attr = pointer.replace('/data/attributes/', '').toLowerCase();
    const map = {
      stores: { name: 'store-name', address: 'store-address' },
      pets: { name: 'pet-name', species: 'pet-species', store: 'pet-store' },
    };
    const ids = map[resource];
    return ids ? ids[attr] || null : null;
  }
  if (pointer.startsWith('/data/relationships/')) {
    const rel = pointer.replace('/data/relationships/', '').toLowerCase();
    if (resource === 'pets' && rel === 'store') return 'pet-store';
  }
  return null;
}

function showValidationErrors(errors, resource, formContainerId) {
  const container = document.getElementById(formContainerId);
  if (!container) return;
  const errorEls = container.querySelectorAll('.field-error');
  errorEls.forEach(el => { el.textContent = ''; });
  const summary = document.getElementById(formContainerId.replace('-form', '-form-errors'));
  if (summary) summary.textContent = '';

  if (!errors || !errors.length) return;

  const byPointer = {};
  errors.forEach(e => {
    const ptr = e.source && e.source.pointer ? e.source.pointer : '';
    if (!byPointer[ptr]) byPointer[ptr] = [];
    byPointer[ptr].push(e.detail || e.title || 'Validation error');
  });

  Object.keys(byPointer).forEach(pointer => {
    const fieldId = pointerToFieldId(pointer, resource);
    const messages = byPointer[pointer];
    if (fieldId) {
      const errEl = document.getElementById(fieldId + '-error');
      if (errEl) errEl.textContent = messages.join('. ');
    }
  });

  const general = errors.filter(e => !e.source || !e.source.pointer || !pointerToFieldId(e.source.pointer, resource));
  if (summary && general.length) {
    summary.textContent = general.map(e => e.detail || e.title).join(' ');
  }
}

async function api(method, path, body = null) {
  const opts = { method, headers: { ...headers } };
  if (body) opts.body = JSON.stringify(body);
  const res = await fetch(API_BASE + path, opts);
  const text = await res.text();
  let data = null;
  try {
    data = text ? JSON.parse(text) : null;
  } catch (_) {}
  if (!res.ok) {
    const err = new Error(data && data.errors ? data.errors[0].detail : res.statusText);
    err.status = res.status;
    err.data = data;
    throw err;
  }
  return data;
}

// --- Stores ---
let stores = [];

async function loadStores() {
  const listEl = document.getElementById('stores-list');
  const errEl = document.getElementById('stores-errors');
  listEl.innerHTML = '';
  errEl.textContent = '';
  try {
    const doc = await api('GET', '/stores');
    stores = (doc && doc.data) ? doc.data : [];
    const ul = document.createElement('ul');
    stores.forEach(s => {
      const li = document.createElement('li');
      const attrs = s.attributes || {};
      li.innerHTML = `
        <span class="item-id">${s.id}</span>
        <span>${escapeHtml(attrs.name || '')}</span>
        <span class="item-actions">
          <button type="button" class="edit-store" data-id="${s.id}">Edit</button>
          <button type="button" class="danger delete-store" data-id="${s.id}">Delete</button>
        </span>`;
      ul.appendChild(li);
    });
    listEl.appendChild(ul);
    listEl.querySelectorAll('.edit-store').forEach(btn => btn.addEventListener('click', () => editStore(btn.dataset.id)));
    listEl.querySelectorAll('.delete-store').forEach(btn => btn.addEventListener('click', () => deleteStore(btn.dataset.id)));
  } catch (e) {
    errEl.textContent = e.message || 'Failed to load stores';
  }
  refreshPetStoreSelect();
}

function refreshPetStoreSelect() {
  const sel = document.getElementById('pet-store');
  if (!sel) return;
  const first = sel.querySelector('option');
  sel.innerHTML = '';
  if (first) sel.appendChild(first);
  stores.forEach(s => {
    const attrs = s.attributes || {};
    const opt = document.createElement('option');
    opt.value = s.id;
    opt.textContent = attrs.name || s.id;
    sel.appendChild(opt);
  });
}

function editStore(id) {
  const s = stores.find(x => x.id === id);
  if (!s) return;
  const attrs = s.attributes || {};
  document.getElementById('store-id').value = id;
  document.getElementById('store-name').value = attrs.name || '';
  document.getElementById('store-address').value = attrs.address || '';
  document.getElementById('store-form-title').textContent = 'Edit store';
  document.getElementById('store-submit').textContent = 'Update';
  document.getElementById('store-cancel').style.display = 'inline-block';
  showValidationErrors([], 'stores', 'store-form');
}

function cancelStoreForm() {
  document.getElementById('store-id').value = '';
  document.getElementById('store-form').reset();
  document.getElementById('store-form-title').textContent = 'Create store';
  document.getElementById('store-submit').textContent = 'Create';
  document.getElementById('store-cancel').style.display = 'none';
  showValidationErrors([], 'stores', 'store-form');
}

async function deleteStore(id) {
  try {
    await api('DELETE', '/stores/' + id);
    await loadStores();
  } catch (e) {
    document.getElementById('stores-errors').textContent = e.message || 'Delete failed';
  }
}

document.getElementById('stores-refresh').addEventListener('click', loadStores);
document.getElementById('store-cancel').addEventListener('click', cancelStoreForm);

document.getElementById('store-form').addEventListener('submit', async (e) => {
  e.preventDefault();
  const id = document.getElementById('store-id').value;
  const name = document.getElementById('store-name').value.trim();
  const address = document.getElementById('store-address').value.trim();
  const body = {
    data: {
      type: 'stores',
      attributes: { name, address },
    },
  };
  if (id) body.data.id = id;
  try {
    showValidationErrors([], 'stores', 'store-form');
    if (id) {
      await api('PATCH', '/stores/' + id, body);
    } else {
      await api('POST', '/stores', body);
    }
    cancelStoreForm();
    await loadStores();
  } catch (err) {
    if (err.data && err.data.errors) {
      showValidationErrors(err.data.errors, 'stores', 'store-form');
    } else {
      document.getElementById('store-form-errors').textContent = err.message || 'Request failed';
    }
  }
});

// --- Pets ---
let pets = [];

async function loadPets() {
  const listEl = document.getElementById('pets-list');
  const errEl = document.getElementById('pets-errors');
  listEl.innerHTML = '';
  errEl.textContent = '';
  try {
    const doc = await api('GET', '/pets');
    pets = (doc && doc.data) ? doc.data : [];
    const ul = document.createElement('ul');
    pets.forEach(p => {
      const attrs = p.attributes || {};
      const storeId = (p.relationships && p.relationships.store && p.relationships.store.data) ? p.relationships.store.data.id : '';
      const storeName = storeId ? (stores.find(s => s.id === storeId)?.attributes?.name || storeId) : '—';
      const li = document.createElement('li');
      li.innerHTML = `
        <span class="item-id">${p.id}</span>
        <span>${escapeHtml(attrs.name || '')}</span>
        <span>${escapeHtml(attrs.species || '')}</span>
        <span class="muted">${escapeHtml(storeName)}</span>
        <span class="item-actions">
          <button type="button" class="edit-pet" data-id="${p.id}">Edit</button>
          <button type="button" class="danger delete-pet" data-id="${p.id}">Delete</button>
        </span>`;
      ul.appendChild(li);
    });
    listEl.appendChild(ul);
    listEl.querySelectorAll('.edit-pet').forEach(btn => btn.addEventListener('click', () => editPet(btn.dataset.id)));
    listEl.querySelectorAll('.delete-pet').forEach(btn => btn.addEventListener('click', () => deletePet(btn.dataset.id)));
  } catch (e) {
    errEl.textContent = e.message || 'Failed to load pets';
  }
}

function editPet(id) {
  const p = pets.find(x => x.id === id);
  if (!p) return;
  const attrs = p.attributes || {};
  const storeId = (p.relationships && p.relationships.store && p.relationships.store.data) ? p.relationships.store.data.id : '';
  document.getElementById('pet-id').value = id;
  document.getElementById('pet-name').value = attrs.name || '';
  document.getElementById('pet-species').value = attrs.species || '';
  document.getElementById('pet-store').value = storeId || '';
  document.getElementById('pet-form-title').textContent = 'Edit pet';
  document.getElementById('pet-submit').textContent = 'Update';
  document.getElementById('pet-cancel').style.display = 'inline-block';
  showValidationErrors([], 'pets', 'pet-form');
}

function cancelPetForm() {
  document.getElementById('pet-id').value = '';
  document.getElementById('pet-form').reset();
  document.getElementById('pet-form-title').textContent = 'Create pet';
  document.getElementById('pet-submit').textContent = 'Create';
  document.getElementById('pet-cancel').style.display = 'none';
  showValidationErrors([], 'pets', 'pet-form');
}

async function deletePet(id) {
  try {
    await api('DELETE', '/pets/' + id);
    await loadPets();
  } catch (e) {
    document.getElementById('pets-errors').textContent = e.message || 'Delete failed';
  }
}

document.getElementById('pets-refresh').addEventListener('click', loadPets);
document.getElementById('pet-cancel').addEventListener('click', cancelPetForm);

document.getElementById('pet-form').addEventListener('submit', async (e) => {
  e.preventDefault();
  const id = document.getElementById('pet-id').value;
  const name = document.getElementById('pet-name').value.trim();
  const species = document.getElementById('pet-species').value;
  const storeId = document.getElementById('pet-store').value;
  const body = {
    data: {
      type: 'pets',
      attributes: { name, species },
    },
  };
  if (id) body.data.id = id;
  if (storeId) {
    body.data.relationships = {
      store: { data: { type: 'stores', id: storeId } },
    };
  }
  try {
    showValidationErrors([], 'pets', 'pet-form');
    if (id) {
      await api('PATCH', '/pets/' + id, body);
    } else {
      await api('POST', '/pets', body);
    }
    cancelPetForm();
    await loadPets();
  } catch (err) {
    if (err.data && err.data.errors) {
      showValidationErrors(err.data.errors, 'pets', 'pet-form');
    } else {
      document.getElementById('pet-form-errors').textContent = err.message || 'Request failed';
    }
  }
});

function escapeHtml(s) {
  const div = document.createElement('div');
  div.textContent = s;
  return div.innerHTML;
}

// Initial load (stores first so pet store dropdown is filled)
loadStores().then(loadPets);
