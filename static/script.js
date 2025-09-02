document.getElementById("searchBtn").addEventListener("click", async () => {
  const uid = document.getElementById("orderUid").value.trim();
  const resultEl = document.getElementById("result");

  if (!uid) {
    resultEl.textContent = "Введите OrderUID!";
    return;
  }

  try {
    const res = await fetch(`/order/${uid}`);
    if (!res.ok) {
      resultEl.textContent = `Ошибка: ${res.status}`;
      return;
    }

    const data = await res.json();
    resultEl.textContent = JSON.stringify(data, null, 2);
  } catch (err) {
    resultEl.textContent = "Ошибка запроса: " + err.message;
  }
});
