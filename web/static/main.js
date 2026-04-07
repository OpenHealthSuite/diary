// OpenFoodDiary - Main JavaScript
// Minimal JS for modal handling and tutorial state

// Modal close helper
function closeModal(event) {
  if (event && event.target !== event.currentTarget) return;
  document.getElementById('modal-container').innerHTML = '';
}

// Close modal on Escape key
document.addEventListener('keydown', (e) => {
  if (e.key === 'Escape') closeModal();
});

// Tutorial localStorage handling
const TUTORIAL_VERSION = 'v1.0';
const TUTORIAL_STORAGE_KEY = 'tutorial-viewed';

function markTutorialSeen() {
  localStorage.setItem(TUTORIAL_STORAGE_KEY, TUTORIAL_VERSION);
}

function shouldShowTutorial() {
  return localStorage.getItem(TUTORIAL_STORAGE_KEY) === null;
}

// Show tutorial on first visit (only on home page)
document.addEventListener('DOMContentLoaded', () => {
  if (shouldShowTutorial() && window.location.pathname === '/') {
    htmx.ajax('GET', '/tutorial', {target: '#modal-container'});
  }
});

// Bulk upload CSV parsing with papaparse
function handleBulkUpload(inputElement) {
  const files = inputElement.files;
  if (!files || files.length === 0) return;

  const statusEl = document.getElementById('upload-status');
  const uploadBtn = document.getElementById('upload-btn');

  statusEl.innerHTML = '<div>Processing...</div>';
  uploadBtn.disabled = true;

  let allLogs = [];
  let processed = 0;

  Array.from(files).forEach(file => {
    file.text().then(strdata => {
      const records = Papa.parse(strdata, {
        header: true,
        skipEmptyLines: true,
      }).data.map(x => {
        const log = {
          ...x,
          time: {
            start: x.timeStart,
            end: x.timeEnd
          },
          metrics: x.metrics ? JSON.parse(x.metrics) : {},
          labels: x.labels ? JSON.parse(x.labels) : []
        };
        delete log.timeStart;
        delete log.timeEnd;
        return log;
      });

      allLogs = [...allLogs, ...records];
      processed++;

      if (processed === files.length) {
        window.bulkUploadLogs = allLogs;
        statusEl.innerHTML = `
          <ul>
            <li>${allLogs.length} logs to upload</li>
          </ul>
        `;
        uploadBtn.disabled = false;
      }
    });
  });
}

async function executeBulkUpload() {
  const logs = window.bulkUploadLogs || [];
  const statusEl = document.getElementById('upload-status');

  let uploaded = 0;
  let errors = 0;

  for (const log of logs) {
    try {
      const method = log.id ? 'PUT' : 'POST';
      const url = log.id ? `/api/logs/${log.id}` : '/api/logs';

      const response = await fetch(url, {
        method: method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(log)
      });

      if (response.ok) {
        uploaded++;
      } else if (response.status == 404) {
        const response = await fetch('/api/logs', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(log)
        });

        if (response.ok) {
          uploaded++;
        } else {

        errors++;
        }

      } else {
        errors++;
      }
    } catch (e) {
      errors++;
    }

    statusEl.innerHTML = `
      <ul>
        <li>Uploaded ${uploaded}/${logs.length}</li>
        <li>Errors: ${errors}</li>
      </ul>
    `;
  }

  if (uploaded === logs.length) {
    statusEl.innerHTML = '<h3>Done!</h3>' + statusEl.innerHTML;
  }
}
