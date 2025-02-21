// ignore_for_file: prefer_const_constructors, prefer_const_literals_to_create_immutables

import 'dart:math';
import 'package:flutter/material.dart';
import 'package:flutter_frontend/resources/app_style.dart';

class FiftyFiftyWidget extends StatelessWidget {
  const FiftyFiftyWidget({
    super.key,
  });

  @override
  Widget build(BuildContext context) {
    return Random().nextBool()
    ? Row(
      children: [
        Icon(
          Icons.keyboard_double_arrow_up,
        ),
        Text(
          '12.2%',
          style: MyAppStyle.regularFontLightGrey,
        )
      ],
    )
    : Row(
      children: [
        Icon(
          Icons.keyboard_double_arrow_down,
        ),
        Text(
          '8.5%',
          style: MyAppStyle.regularFontLightGrey,
        )
      ],
    );
  }
}