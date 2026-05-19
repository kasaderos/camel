import sys
import re

def main():
    print("Вставьте данные (сумму и веса) и нажмите Ctrl+D (Linux/Mac) или Ctrl+Z (Windows) + Enter для обработки:\n")

    # Читаем все строки из стандартного ввода
    try:
        input_data = sys.stdin.read()
    except EOFError:
        pass

    if not input_data.strip():
        print("Ошибка: Данные не введены.")
        return

    # 1. Извлекаем общую сумму (первое число в тексте)
    budget_match = re.search(r'(\d+)', input_data)
    if not budget_match:
        print("Ошибка: Сумма в начале текста не найдена.")
        return

    total_budget = float(budget_match.group(1))

    # 2. Ищем все вхождения asset_id и weight
    # Паттерн учитывает возможные пробелы и разные форматы написания
    pattern = r"asset_id=([\w.]+)\s+weight=([\d.]+)"
    assets = re.findall(pattern, input_data)

    if not assets:
        print("Ошибка: Не удалось найти данные по активам и весам.")
        return

    # 3. Вывод результатов
    print(f"\n{'='*40}")
    print(f"{'АКТИВ':<15} | {'СУММА':>15}")
    print(f"{'-'*40}")

    calculated_total = 0
    for ticker, weight in assets:
        amount = total_budget * float(weight)
        calculated_total += amount
        print(f"{ticker:<15} | {amount:>15,.2f}")

    print(f"{'-'*40}")
    print(f"{'ИТОГО:':<15} | {calculated_total:>15,.2f}")
    print(f"{'='*40}")

if __name__ == "__main__":
    main()
