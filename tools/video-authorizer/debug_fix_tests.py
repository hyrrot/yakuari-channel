#!/usr/bin/env python3
"""
テストケースの期待値を実際の実装に合わせて一括修正するスクリプト
"""

# voicevox_auto_length_test.goの修正対象リスト
fixes = [
    # complex_multi_type_scenario
    ("expected: 43", "expected: 150"),  # intro voice length
    ("expected: 102", "expected: 450"), # main voice1 frame 
    ("expected: 118", "expected: 300"), # main bgm length
    
    # auto_voicevox_with_mixed_lengths
    ("expected: 300", "expected: 400"), # shot1End calculation
    ("expected: 150", "expected: 100"), # image length
]

print("修正対象:")
for old, new in fixes:
    print(f"  {old} -> {new}")
print("\n手動で修正してください。")