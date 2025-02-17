// TODO: remove these ignores for when this class is finished
// ignore_for_file: prefer_const_constructors, prefer_const_literals_to_create_immutables

import 'package:flutter/material.dart';
import 'package:flutter_frontend/resources/app_style.dart';

class PortfolioCard extends StatelessWidget {
  const PortfolioCard({
    super.key,
  });

  @override
  Widget build(BuildContext context) {

    // TODO: replace this with a function call to API
    List<String> listData = [
      'One',
      'Two',
      'Three',
      'Foiur',
    ];

    return Expanded(
      child: Card(
        child: Column(
          children: [
            Text(
              'Portfolio',
              style: MyAppStyle.largeFont
            ),
            Expanded(
              child: ListView.builder(
                itemCount: listData.length,
                itemBuilder: (context, index) {
                  return PortfolioItem(itemData: listData[index]);
                }
              )
            ),
            ElevatedButton(
              onPressed: () => {},
              child: Text(
                'Search the Market',
                style: MyAppStyle.regularFont,
              )
            ),
            SizedBox(
              height: 8.0,
            )
          ],
        ),
      ),
    );
  }
}

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
