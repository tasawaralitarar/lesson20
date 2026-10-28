async function fetchOrders() {
  try {
    const response = await fetch('/api/orders');
    const orders = await response.json();

    const container = document.getElementById('orders-list');
    if (!orders || orders.length === 0) {
      container.innerHTML = '<p>注文履歴はありません。</p>';
      return;
    }

    let html = `
      <table>
        <thead>
          <tr>
            <th>注文ID</th>
            <th>商品名</th>
            <th>単価</th>
            <th>数量</th>
            <th>合計金額</th>
            <th>注文日時</th>
            <th>決済</th>
          </tr>
        </thead>
        <tbody>
    `;

    orders.forEach(order => {
      html += `
        <tr>
          <td>${order.id}</td>
          <td>${order.product_name}</td>
          <td>¥${order.price.toLocaleString()}</td>
          <td>${order.quantity}</td>
          <td>¥${order.total_price.toLocaleString()}</td>
          <td>${new Date(order.created_at).toLocaleString()}</td>
          <td>
            <a href="payment.html?order_id=${order.id}&amount=${order.total_price}" class="btn">決済する</a>
          </td>
        </tr>
      `;
    });

    html += `
        </tbody>
      </table>
    `;

    container.innerHTML = html;
  } catch (error) {
    console.error('Error:', error);
  }
}

fetchOrders();