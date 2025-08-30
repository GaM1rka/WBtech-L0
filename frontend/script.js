const API_BASE_URL = 'http://localhost:8081';

function searchOrder() {
    const orderId = document.getElementById('orderIdInput').value.trim();
    
    if (!orderId) {
        showError('Пожалуйста, введите ID заказа');
        return;
    }

    // Проверка на минимальную длину UUID
    if (orderId.length < 10) {
        showError('ID заказа слишком короткий');
        return;
    }

    hideError();
    showLoading();
    hideResult();

    fetch(`${API_BASE_URL}/order/${orderId}`)
        .then(response => {
            if (!response.ok) {
                if (response.status >= 400) {
                    throw new Error('Заказ не найден');
                } else if (response.status === 400) {
                    throw new Error('Неверный формат ID заказа');
                } else {
                    throw new Error(`Ошибка сервера: ${response.status}`);
                }
            }
            return response.json();
        })
        .then(data => {
            displayResult(data);
        })
        .catch(error => {
            showError(error.message || 'Произошла ошибка при получении данных');
        })
        .finally(() => {
            hideLoading();
        });
}

function displayResult(data) {
    const resultElement = document.getElementById('orderData');
    resultElement.textContent = JSON.stringify(data, null, 2);
    showResult();
}

function showLoading() {
    document.getElementById('loading').classList.remove('hidden');
}

function hideLoading() {
    document.getElementById('loading').classList.add('hidden');
}

function showResult() {
    document.getElementById('result').classList.remove('hidden');
}

function hideResult() {
    document.getElementById('result').classList.add('hidden');
}

function showError(message) {
    const errorElement = document.getElementById('error');
    errorElement.textContent = message;
    errorElement.classList.remove('hidden');
    hideResult();
}

function hideError() {
    document.getElementById('error').classList.add('hidden');
}

function handleKeyPress(event) {
    if (event.key === 'Enter') {
        searchOrder();
    }
}

// Очистка ошибки при начале ввода
document.getElementById('orderIdInput').addEventListener('input', hideError);