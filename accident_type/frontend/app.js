document.addEventListener('DOMContentLoaded', function() {
    const API_BASE_URL = 'http://localhost:8080';
    const searchBtn = document.getElementById('search-btn');
    const resultsContainer = document.getElementById('results-container');

    // Set default dates
    const endDate = new Date();
    const startDate = new Date();
    startDate.setMonth(endDate.getMonth() - 3);
    document.getElementById('start-date').value = startDate.toISOString().split('T')[0];
    document.getElementById('end-date').value = endDate.toISOString().split('T')[0];

    searchBtn.addEventListener('click', async function() {
        const startDate = document.getElementById('start-date').value;
        const endDate = document.getElementById('end-date').value;
        const accidentType = document.getElementById('accident-type').value.trim();

        if (!startDate || !endDate) {
            alert('Please select both start and end dates');
            return;
        }

        try {
            // Show loading state
            searchBtn.disabled = true;
            resultsContainer.innerHTML = '<p>Loading...</p>';

            // Make API request
            const url = new URL(`${API_BASE_URL}/accidents`);
            url.searchParams.append('start_date', startDate);
            url.searchParams.append('end_date', endDate);
            if (accidentType) url.searchParams.append('accident_type', accidentType);

            const response = await fetch(url);
            
            if (!response.ok) {
                const error = await response.json().catch(() => ({}));
                throw new Error(error.details || `Server error: ${response.status}`);
            }

            const data = await response.json();
            displayResults(data);
        } catch (error) {
            console.error('API Error:', error);
            resultsContainer.innerHTML = `
                <div class="error">
                    ${error.message}<br>
                    <small>Ensure backend is running at ${API_BASE_URL}</small>
                </div>`;
        } finally {
            searchBtn.disabled = false;
        }
    });

    function displayResults(logs) {
        if (!logs || !logs.length) {
            resultsContainer.innerHTML = '<p class="no-results">No matching records found</p>';
            return;
        }

        resultsContainer.innerHTML = logs.map(log => `
            <div class="log-entry">
                <h3>${log.AccidentType || 'Unknown Type'}</h3>
                <p><strong>Date:</strong> ${log.Date}</p>
                <p><strong>Reported by:</strong> ${log.ReportedBy}</p>
                <p><strong>Details:</strong> ${log.Comments}</p>
            </div>
        `).join('');
    }
});