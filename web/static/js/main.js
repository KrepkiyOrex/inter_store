// Функция для отправки данных на сервер
function updateCart(productID, quantity) {
    fetch('/update_cart', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: new URLSearchParams({
            'product_id': productID,
            'quantity': quantity
        })
    }).then(response => response.json())
      .then(data => {
          // Обновляем общую стоимость для каждой позиции
          document.querySelectorAll('.quantity-input').forEach(input => {
              const productID = input.dataset.productId;
              const item = data.cart_items.find(item => item.product_id == productID);
              if (item) {
                  const totalElement = input.closest('tr').querySelector('.total-price');
                  const itemTotal = (item.price * item.quantity).toFixed(2);
                  totalElement.textContent = `$ ${itemTotal}`;
              }
          });

          // Обновляем общую сумму
          const totalSumElement = document.querySelector('.total-sum');
          totalSumElement.textContent = `$ ${data.total_sum.toFixed(2)}`;
      })
      .catch(error => {
          console.error('Error:', error);
      });
}

// Функция для обработки отправки формы
function submitOrder(event) {
    event.preventDefault(); // Отменяем стандартную отправку формы

    const form = event.target;
    const formData = new FormData(form);

    fetch('/submit_order', {
        method: 'POST',
        body: new URLSearchParams(formData)
    }).then(response => {
        if (response.ok) {
            window.location.href = '/users-orders'; // Перенаправление при успешной отправке
        } else {
            console.error('Failed to submit order');
        }
    }).catch(error => {
        console.error('Error:', error);
    });
}

// Инициализация обработчиков после загрузки DOM
document.addEventListener('DOMContentLoaded', () => {
    document.querySelectorAll('.quantity-input').forEach(input => {
        input.addEventListener('input', function() {
            const productID = this.dataset.productId;
            const quantity = this.value;

            console.log(`Updating product ${productID} with quantity ${quantity}`);
            updateCart(productID, quantity); // Отправляем обновление на сервер
        });
    });

    // Инициализируем обработчик формы
    document.getElementById('cart-form').addEventListener('submit', submitOrder);

    // Первоначальный расчет суммы
    updateTotal();
});

function updateTotal() {
    let totalSum = 0;

    document.querySelectorAll('.quantity-input').forEach(input => {
        const quantity = input.value;
        const priceElement = input.closest('tr').querySelector('.price');
        const totalElement = input.closest('tr').querySelector('.total-price');

        const price = parseFloat(priceElement.textContent.replace('$', ''));
        const totalPrice = price * quantity;

        totalElement.textContent = `$ ${totalPrice.toFixed(2)}`;

        totalSum += totalPrice;
    });

    document.querySelector('.total-sum').textContent = `$ ${totalSum.toFixed(2)}`;
}
