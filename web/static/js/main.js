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
            }).then(response => {
                if (!response.ok) {
                    console.error('Failed to update cart');
                    return;
                }
                return response.text(); // Получаем текстовый ответ
            }).then(data => {
                if (data) {
                    try {
                        const jsonData = JSON.parse(data);
                        console.log('Response JSON:', jsonData);
                    } catch (error) {
                        console.error('Invalid JSON:', error);
                    }
                } else {
                    console.warn('Empty response');
                }
            }).catch(error => {
                console.error('Error:', error);
            });
        }

        // Функция для обновления общей стоимости позиции и суммарной стоимости
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

        // Добавляем обработчик события для каждого поля количества
        document.addEventListener('DOMContentLoaded', () => {
            document.querySelectorAll('.quantity-input').forEach(input => {
                input.addEventListener('input', function() {
                    const productID = this.dataset.productId;
                    const quantity = this.value;

                    console.log(`Updating product ${productID} with quantity ${quantity}`);
                    updateCart(productID, quantity); // Отправляем обновление на сервер
                    updateTotal(); // Обновляем сумму
                });
            });

            // Первоначальный расчет суммы
            updateTotal();
        });

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