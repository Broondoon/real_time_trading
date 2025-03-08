// ignore_for_file: prefer_const_constructors, prefer_const_literals_to_create_immutables

import 'package:flutter/material.dart';
import 'package:flutter_frontend/resources/app_style.dart';

class HistoryItem extends StatelessWidget {
  const HistoryItem({
    super.key,
    required this.totalPrice,
    required this.stockPrice,
    required this.quantity,
    required this.timestamp,
  });

  final int totalPrice;
  final int stockPrice;
  final int quantity;
  final String timestamp;

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        Card(
          child: SizedBox(
            height: 40,
            child: Padding(
              padding: const EdgeInsets.fromLTRB(8.0, 0, 8.0, 0),
              child: Row(
                children: [
                  Text(
                    'Wallet Change: $totalPrice',
                    style: MyAppStyle.regularFont,
                  ),
                  VerticalDivider(),
                  Text(
                    '$quantity stock traded for $stockPrice @ $timestamp',
                    // style: MyAppStyle.regularFont,
                  ),
                ],
              ),
            ),
          ),
        ),
      ],
    );
  }
}