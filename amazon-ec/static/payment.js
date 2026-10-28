document.addEventListener('DOMContentLoaded', () => {
  const urlParams = new URLSearchParams(window.location.search);
  const orderId = urlParams.get('order_id');
  const amount = urlParams.get('amount');

  const displayOrderId = document.getElementById('display-order-id');
  const displayAmount = document.getElementById('display-amount');
  const paymentForm = document.getElementById('payment-form');
  const paymentMessage = document.getElementById('payment-message');

  if (orderId) {
    displayOrderId.textContent = orderId;
  }
  if (amount) {
    displayAmount.textContent = Number(amount).toLocaleString();
  }

  paymentForm.addEventListener('submit', async (e) => {
    e.preventDefault();

    if (!orderId || !amount) {
      paymentMessage.textContent = '注文情報が正しく取得できませんでした。';
      paymentMessage.style.color = 'red';
      return;
    }

    const method = document.getElementById('method').value;

    try {
      const response = await fetch('/api/payments', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          order_id: parseInt(orderId, 10),
          amount: parseInt(amount, 10),
          method: method
        })
      });

      if (response.ok) {
        paymentMessage.textContent = '決済が正常に完了しました。';
        paymentMessage.style.color = 'green';
        setTimeout(() => {
          window.location.href = 'orders.html';
        }, 1500);
      } else {
        const errData = await response.json();
        paymentMessage.textContent = '決済処理に失敗しました: ' + (errData.message || response.statusText);
        paymentMessage.style.color = 'red';
      }
    } catch (error) {
      console.error('Error:', error);
      paymentMessage.textContent = '通信エラーが発生しました。';
      paymentMessage.style.color = 'red';
    }
  });
});