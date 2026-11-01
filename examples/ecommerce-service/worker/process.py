import time

def process_order(order_id: str):
    print(f"Processing order {order_id}...")
    validate_order(order_id)

def validate_order(order_id: str) -> bool:
    return len(order_id) > 0

if __name__ == "__main__":
    process_order("ord_999")
