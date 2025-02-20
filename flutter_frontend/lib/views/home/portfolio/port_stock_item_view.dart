// TODO: remove these ignores for when this class is finished
// ignore_for_file: prefer_const_constructors, prefer_const_literals_to_create_immutables

import 'package:flutter/material.dart';
import 'package:flutter_frontend/resources/app_style.dart';

class PortfolioItem extends StatelessWidget {
  const PortfolioItem({
    super.key,
    required this.stockName,
    required this.stockPrice,
    required this.quantOwned,
    required this.isPending,
  });

  final String stockName;
  final String stockPrice;
  final String quantOwned;
  final bool isPending;

  Widget showPending() {
    if (isPending) {
      return Card(
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
      );
    }
    else {
      // Basically return nothing, but because of null typing,
      //    we have to return SOMETHING.
      return SizedBox();
    }
  }

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
                    '${quantOwned}x $stockName',
                    style: MyAppStyle.regularFont,
                  ),
                  // TODO: we don't get stock prices from the getPortfolio call; later for bonus points, ask for stock prices somehow
                  // VerticalDivider(),
                  // Text(
                  //   '\$$stockPrice',
                  //   style: MyAppStyle.regularFont,
                  // ),
                ],
              ),
            ),
          ),
        ),
        showPending(),
      ],
    );
  }
}
