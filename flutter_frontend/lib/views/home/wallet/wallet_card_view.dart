// ignore_for_file: prefer_const_constructors, prefer_const_literals_to_create_immutables

import 'package:flutter/material.dart';
import 'package:flutter_frontend/resources/app_style.dart';

class WalletCard extends StatelessWidget {
  const WalletCard({
    super.key,
  });

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: Card(
        child: Center(
          child: Column(
            children: [
              Text(
                'Wallet',
                style: MyAppStyle.largeFont
              ),
            ],
          )
        ),
      ),
    );
  }
}
