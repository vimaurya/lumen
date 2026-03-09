import asyncio
import aiohttp
import random
import time

TARGET_URL = "http://localhost:8080/api/v1/34"
NUM_USERS = 10          
REQUESTS_PER_USER = 1000
CONCURRENCY = 50

def generate_fake_ip():
    return f"{random.randint(1, 255)}.{random.randint(1, 255)}.{random.randint(1, 255)}.{random.randint(1, 255)}"

async def simulate_user(session, user_id):
    fake_ip = generate_fake_ip()
    headers = {
        "X-Forwarded-For": fake_ip,
        "User-Agent": f"LumenLoadTester/1.0 (User-{user_id})"
    }
    
    success_count = 0
    blocked_count = 0
    
    for _ in range(REQUESTS_PER_USER):
        try:
            async with session.get(TARGET_URL, headers=headers) as response:
                if response.status == 200:
                    success_count += 1
                elif response.status == 429:
                    blocked_count += 1
        except Exception as e:
            print(f"User {user_id} error: {e}")
        
    return user_id, fake_ip, success_count, blocked_count

async def main():
    print(f"Starting simulation: {NUM_USERS} users hitting {TARGET_URL}...")
    start_time = time.perf_counter()
    
    connector = aiohttp.TCPConnector(limit=CONCURRENCY)
    async with aiohttp.ClientSession(connector=connector) as session:
        tasks = [simulate_user(session, i) for i in range(NUM_USERS)]
        results = await asyncio.gather(*tasks)
    
    end_time = time.perf_counter()
    duration = end_time - start_time
    
    print("\n--- Simulation Summary ---")
    total_200 = sum(r[2] for r in results)
    total_429 = sum(r[3] for r in results)
    print(f"Total Time: {duration:.2f}s")
    print(f"Total Requests: {NUM_USERS * REQUESTS_PER_USER}")
    print(f"Successful (200 OK): {total_200}")
    print(f"Rate Limited (429): {total_429}")
    print(f"Avg RPS: {(total_200 + total_429) / duration:.2f}")

if __name__ == "__main__":
    asyncio.run(main())
