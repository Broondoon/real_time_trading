// TODO: remove these ignores for when this class is finished
// ignore_for_file: prefer_const_constructors, prefer_const_literals_to_create_immutables

import 'package:flutter/material.dart';
import 'package:flutter_frontend/resources/app_style.dart';

class PortfolioItem extends StatelessWidget {
  const PortfolioItem({
    super.key,
    required this.itemData,
  });

  final String itemData;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Card(
          child: SizedBox(
            height: 40,
            child: Padding(
              padding: const EdgeInsets.fromLTRB(8.0, 0, 8.0, 0),
              child: Row(
                children: [
                  Text(
                    '10x GOOG',
                    style: MyAppStyle.regularFont,
                  ),
                  VerticalDivider(),
                  Text(
                    '\$999.99',
                    style: MyAppStyle.regularFont,
                  ),
                ],
              ),
            ),
          ),
        ),
        Card(
          child: SizedBox(
            height: 40,
            child: Padding(
              padding: const EdgeInsets.fromLTRB(8.0, 0, 8.0, 0),
              child: Align(
                alignment: Alignment.centerLeft,
                child: Text(
                  'Pending...',
                  style: MyAppStyle.regularFont,
                ),
              ),
            ),
          ),
        ),
      ],
    );
  }
}
